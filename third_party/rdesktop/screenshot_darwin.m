#import <AppKit/AppKit.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>

void get_cursor_size(int *width, int *height) {
    NSSize size = [[[NSCursor currentSystemCursor] image] size];
    *width = size.width;
    *height = size.height;
}

static int copy_image(CGImageRef image, unsigned char *pixels, int width, int height) {
    if (!image || width <= 0 || height <= 0) return -1;
    CGColorSpaceRef space = CGColorSpaceCreateDeviceRGB();
    CGContextRef context = CGBitmapContextCreate(pixels, width, height, 8, width * 4,
        space, kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big);
    CGColorSpaceRelease(space);
    if (!context) return -1;
    CGContextDrawImage(context, CGRectMake(0, 0, width, height), image);
    CGContextRelease(context);
    return 0;
}

void cursor_copy(unsigned char *pixels, int width, int height) {
    @autoreleasepool {
        NSImage *image = [[NSCursor currentSystemCursor] image];
        copy_image([image CGImageForProposedRect:nil context:nil hints:nil], pixels, width, height);
    }
}

int screenshot(unsigned char *pixels, int width, int height, bool show_cursor) {
    @autoreleasepool {
        if (@available(macOS 14.0, *)) {
            // Async callbacks never retain a pointer into Go memory.
            dispatch_semaphore_t ready = dispatch_semaphore_create(0);
            __block SCShareableContent *content = nil;
            [SCShareableContent getShareableContentExcludingDesktopWindows:NO
                onScreenWindowsOnly:YES completionHandler:^(SCShareableContent *value, NSError *error) {
                    content = [value retain];
                    dispatch_semaphore_signal(ready);
                }];
            if (dispatch_semaphore_wait(ready, dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_SEC))) {
                dispatch_async(dispatch_get_global_queue(QOS_CLASS_UTILITY, 0), ^{
                    dispatch_semaphore_wait(ready, DISPATCH_TIME_FOREVER);
                    [content release];
                    dispatch_release(ready);
                });
                return -1;
            }
            SCDisplay *display = nil;
            for (SCDisplay *candidate in content.displays) {
                if (candidate.displayID == CGMainDisplayID()) { display = candidate; break; }
            }
            if (!display) { [content release]; dispatch_release(ready); return -1; }
            SCContentFilter *filter = [[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];
            SCStreamConfiguration *config = [[SCStreamConfiguration alloc] init];
            config.width = width;
            config.height = height;
            config.showsCursor = show_cursor;
            __block CGImageRef captured = NULL;
            [SCScreenshotManager captureImageWithFilter:filter configuration:config
                completionHandler:^(CGImageRef image, NSError *error) {
                    if (image) captured = CGImageRetain(image);
                    dispatch_semaphore_signal(ready);
                }];
            [config release];
            [filter release];
            [content release];
            if (dispatch_semaphore_wait(ready, dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_SEC))) {
                dispatch_async(dispatch_get_global_queue(QOS_CLASS_UTILITY, 0), ^{
                    dispatch_semaphore_wait(ready, DISPATCH_TIME_FOREVER);
                    if (captured) CGImageRelease(captured);
                    dispatch_release(ready);
                });
                return -1;
            }
            int result = copy_image(captured, pixels, width, height);
            if (captured) CGImageRelease(captured);
            dispatch_release(ready);
            return result;
        }
        return -1;
    }
}
