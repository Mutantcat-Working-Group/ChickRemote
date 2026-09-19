var page = {
    init: function() {
        $('#rule-type').change(page.render);
        page.render();
        setInterval(page.render, 5000);
    },
    render: function() {
        page.render_cards();
        page.render_rules();
    },
    render_cards: function() {
        $.get('/api/info', function(ret) {
            $('#cards').empty();
            page.add_card('规则总数', ret.rules);
            page.add_card('虚拟连接数', ret.virtual_links);
            page.add_card('终端会话', ret.sessions);
        });
    },
    render_rules: function() {
        $.get('/api/rules', function(ret) {
            $('#rules tbody').empty();
            var type = $('#rule-type').val();
            $.each(ret, function(_, rule) {
                if (type != 'all' && rule.type != type) {
                    return;
                }
                var send_bytes = 0;
                var send_packet = 0;
                var recv_bytes = 0;
                var recv_packet = 0;
                $.each(rule.links, function(_, link) {
                    send_bytes += link.send_bytes;
                    send_packet += link.send_packet;
                    recv_bytes += link.recv_bytes;
                    recv_packet += link.recv_packet;
                });
                var op = $('<td>');
                switch (rule.type) {
                case 'shell':
                case 'vnc':
                case 'code-server':
                    op.append($('<a>').attr({
                        href: '/terminal.html?name=' + encodeURIComponent(rule.name),
                        target: '_blank',
                        rel: 'noopener'
                    }).text('连接'));
                    break;
                }
                var row = $('<tr>');
                [rule.name, rule.remote, rule.type, rule.links ? rule.links.length : 0,
                    humanize.bytes(recv_bytes) + '/' + humanize.bytes(send_bytes),
                    recv_packet + '/' + send_packet].forEach(function(value) {
                    row.append($('<td>').text(value));
                });
                $('#rules tbody').append(row.append(op));
            });
        });
    },
    add_card: function(title, count) {
        var str = `
        <div class="col-lg-3">
            <div class="card">
                <div class="card-header border-0">
                    <div class="d-flex justify-content-between">
                        <h3 class="card-title">${title}</h3>
                    </div>
                </div>
                <div class="card-body">
                    <div class="d-flex">
                        <p class="d-flex flex-column">
                            <span class="text-bold text-lg">${count}</span>
                        </p>
                    </div>
                </div>
            </div>
        </div>`;
        $('#cards').append($(str));
    }
};
$(document).ready(page.init);
