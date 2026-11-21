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

test('link completion replaces only an authorized, server-rendered card fragment', async () => {
    const attrs = {
        'data-live-event': 'external-link-card',
        'data-live-card': '1-' + 'a'.repeat(64),
        'data-live-url': '/private/topic/1/thread/1/link-preview?card=1-' + 'a'.repeat(64),
        'data-live-swap': 'outerHTML'
    };
    const replacement = {textContent: 'Fetched preview'};
    const template = {content: {firstElementChild: replacement}};
    let inserted;
    const card = {
        getAttribute: name => attrs[name] || null,
        replaceWith: element => { inserted = element; }
    };
    const doc = {
        querySelectorAll: () => [card, threadTarget()],
        createElement: type => {
            assert.equal(type, 'template');
            return template;
        }
    };
    const fetch = async path => {
        assert.equal(path, attrs['data-live-url']);
        return {
            ok: true,
            headers: {get: () => 'text/html'},
            text: async () => '<div class="external-link-card">Fetched preview</div>'
        };
    };
    assert.equal(await live.handleMessage({kind: 'external-link-card', data: 'ignored'}, doc, fetch), 1);
    assert.equal(inserted, replacement);
    assert.equal(template.innerHTML, '<div class="external-link-card">Fetched preview</div>');
});

test('card refresh fan-out uses the same per-page bound as the socket watch', () => {
    const key = '123-' + 'a'.repeat(64);
    const cards = Array.from({length: 20}, () => ({
        getAttribute: name => name === 'data-live-card' ? key : 'external-link-card'
    }));
    const doc = {querySelectorAll: () => cards};
    assert.equal(live.matchingTargets(doc, 'external-link-card').length, 12);
});
