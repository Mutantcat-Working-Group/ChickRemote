use std::process::{Child, Command};

#[derive(Default)]
pub struct Engine {
    child: Option<Child>,
    #[cfg(windows)]
    job: Option<Job>,
}

impl Engine {
    pub fn running(&mut self) -> Result<bool, String> {
        if let Some(child) = &mut self.child {
            if child.try_wait().map_err(|e| e.to_string())?.is_none() {
                return Ok(true);
            }
            self.stop()?;
        }
        Ok(false)
    }

    pub fn start(&mut self, mut command: Command) -> Result<(), String> {
        if self.running()? {
            return Err("Client is already running".into());
        }
        #[cfg(unix)]
        {
            use std::os::unix::process::CommandExt;
            command.process_group(0);
        }
        #[cfg(windows)]
        {
            use std::os::windows::process::CommandExt;
            command.creation_flags(0x08000000); // CREATE_NO_WINDOW
        }
        let child = command
            .spawn()
            .map_err(|e| format!("Cannot start client: {e}"))?;
        self.child = Some(child);
        #[cfg(windows)]
        {
            match Job::attach(self.child.as_ref().unwrap()) {
                Ok(job) => self.job = Some(job),
                Err(error) => {
                    let _ = self.stop();
                    return Err(error);
                }
            }
        }
        Ok(())
    }

    pub fn stop(&mut self) -> Result<(), String> {
        if let Some(mut child) = self.child.take() {
            #[cfg(unix)]
            unsafe {
                libc::kill(-(child.id() as i32), libc::SIGKILL);
            }
            #[cfg(windows)]
            {
                self.job.take();
            }
            let _ = child.kill();
            child.wait().map_err(|e| e.to_string())?;
        }
        Ok(())
    }
}

impl Drop for Engine {
    fn drop(&mut self) {
        let _ = self.stop();
    }
}

#[cfg(windows)]
struct Job(windows_sys::Win32::Foundation::HANDLE);
#[cfg(windows)]
unsafe impl Send for Job {}
#[cfg(windows)]
impl Job {
    fn attach(child: &Child) -> Result<Self, String> {
        use std::os::windows::io::AsRawHandle;
        use windows_sys::Win32::System::JobObjects::*;
        unsafe {
            let handle = CreateJobObjectW(std::ptr::null(), std::ptr::null());
            if handle.is_null() {
                return Err(std::io::Error::last_os_error().to_string());
            }
            let job = Self(handle);
            let mut limits: JOBOBJECT_EXTENDED_LIMIT_INFORMATION = std::mem::zeroed();
            limits.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
            if SetInformationJobObject(
                handle,
                JobObjectExtendedLimitInformation,
                &limits as *const _ as _,
                std::mem::size_of_val(&limits) as u32,
            ) == 0
                || AssignProcessToJobObject(handle, child.as_raw_handle() as _) == 0
            {
                return Err(std::io::Error::last_os_error().to_string());
            }
            Ok(job)
        }
    }
}
#[cfg(windows)]
impl Drop for Job {
    fn drop(&mut self) {
        unsafe {
            windows_sys::Win32::Foundation::CloseHandle(self.0);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn missing_binary_returns_error() {
        let mut engine = Engine::default();
        assert!(engine
            .start(std::process::Command::new("/nonexistent/chickreomte-test"))
            .is_err());
        assert!(!engine.running().unwrap());
    }
    #[test]
    #[cfg(unix)]
    fn prevents_duplicate_start_and_reaps_child() {
        let mut engine = Engine::default();
        let mut command = std::process::Command::new("/bin/sh");
        command.args(["-c", "sleep 60"]);
        engine.start(command).unwrap();
        assert!(engine.running().unwrap());
        assert!(engine.start(std::process::Command::new("/bin/sh")).is_err());
        engine.stop().unwrap();
        assert!(!engine.running().unwrap());
        engine.stop().unwrap();
    }
}
