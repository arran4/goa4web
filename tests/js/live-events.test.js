const assert = require('node:assert/strict');
const test = require('node:test');
const live = require('../../core/templates/assets/live-events.js');

function threadTarget() {
    const attrs = {
        'data-live-event': 'thread-reply',
        'data-live-url': '/private/topic/1/thread/1/live?after=3'
    };
    return {
        innerHTML: '',
        getAttribute: name => attrs[name] || null
    };
}

test('thread signals reach only the thread prompt and reconcile from server HTML', async () => {
    const target = threadTarget();
    const doc = {querySelectorAll: () => [target]};
    let calls = 0;
    const fetch = async (url, options) => {
        assert.equal(url, target.getAttribute('data-live-url'));
        assert.equal(options.credentials, 'same-origin');
        assert.equal(options.headers['X-Live-Fragment'], '1');
        calls++;
        return {
            ok: true,
            status: calls === 1 ? 204 : 200,
            headers: {get: () => 'text/html; charset=utf-8'},
            text: async () => '<p>New replies are available.</p>'
        };
    };

    assert.equal(await live.handleMessage('{"kind":"notification"}', doc, fetch), 0);
    assert.equal(calls, 0);
    assert.equal(await live.refreshAllTargets(doc, fetch), 0);
    assert.equal(target.innerHTML, '');
    assert.equal(await live.handleMessage('{"kind":"thread-reply","data":"<script>bad()</script>"}', doc, fetch), 1);
    assert.equal(target.innerHTML, '<p>New replies are available.</p>');
    assert.equal(await live.refreshAllTargets(doc, fetch), 1);
});

test('live fragment fetch accepts only same-origin paths', () => {
    for (const url of ['//example.com/', '/\\example.com/', '/\texample.com/', 'https://example.com/']) {
        assert.equal(live.isSafeFragmentURL(url), false, url);
    }
    assert.equal(live.isSafeFragmentURL('/private/topic/1/thread/1/live?after=3'), true);
});
