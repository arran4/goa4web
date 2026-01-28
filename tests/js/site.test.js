const fs = require('fs');
const path = require('path');
const { TextEncoder } = require('util');

// Mock DOM
global.Node = {
    TEXT_NODE: 3,
    ELEMENT_NODE: 1
};

global.TextEncoder = TextEncoder;

class MockNode {
    constructor(type, content = "") {
        this.nodeType = type;
        this.textContent = content;
        this.childNodes = [];
        this.parentElement = null;
        this.attributes = {};
    }

    appendChild(child) {
        child.parentElement = this;
        this.childNodes.push(child);
        // Minimal textContent update logic for Element (not fully robust but enough for test)
        if (this.nodeType === Node.ELEMENT_NODE) {
             // In real DOM textContent is concatenation.
             // But calculateSourceOffset accesses child.textContent or node.textContent directly.
        }
    }

    get parentNode() { return this.parentElement; }

    hasAttribute(name) {
        return name in this.attributes;
    }

    getAttribute(name) {
        return this.attributes[name];
    }

    setAttribute(name, value) {
        this.attributes[name] = value;
    }
}

global.document = {
    addEventListener: () => {},
    body: {
        addEventListener: () => {}
    }
};
global.window = global;
global.navigator = { clipboard: { writeText: () => Promise.resolve() } };

// Load site.js
const siteJsPath = path.join(__dirname, '../../core/templates/assets/site.js');
const siteJsContent = fs.readFileSync(siteJsPath, 'utf8');
eval(siteJsContent);

// Tests
function assert(condition, message) {
    if (!condition) {
        console.error("FAIL:", message);
        process.exit(1);
    } else {
        console.log("PASS:", message);
    }
}

console.log("Running JS Tests...");

// Test 1: Simple text in a span with start-pos
// <span data-start-pos="10">Hello</span>
// Select "ll" (offset 2). Expected: 10 + 2 = 12.
const span1 = new MockNode(Node.ELEMENT_NODE);
span1.setAttribute('data-start-pos', '10');
const text1 = new MockNode(Node.TEXT_NODE, "Hello");
span1.appendChild(text1);

const res1 = calculateSourceOffset(text1, 2);
assert(res1 === 12, `Simple text offset. Got ${res1}, want 12`);

// Test 2: Multi-byte characters (Emoji)
// <span data-start-pos="100">😀Hello</span>
// Emoji is 4 bytes.
// Select "H" (offset 2 in UTF-16, because 😀 is 2 chars).
// textContent: "\uD83D\uDE00Hello"
// prefix: "😀" (2 chars).
// byteLen of "😀" is 4.
// Expected: 100 + 4 = 104.
const span2 = new MockNode(Node.ELEMENT_NODE);
span2.setAttribute('data-start-pos', '100');
const text2 = new MockNode(Node.TEXT_NODE, "😀Hello");
span2.appendChild(text2);

// Offset of 'H' in "😀Hello" is 2 (surrogate pair counts as 2)
const res2 = calculateSourceOffset(text2, 2);
assert(res2 === 104, `Emoji offset. Got ${res2}, want 104`);

// Test 3: Element selection (offset into childNodes)
// <div data-start-pos="50"><span data-start-pos="60">Child</span></div>
// Select div at offset 0 (before span).
// Should return start pos of child 0 -> 60.
const div3 = new MockNode(Node.ELEMENT_NODE);
div3.setAttribute('data-start-pos', '50');
const span3 = new MockNode(Node.ELEMENT_NODE);
span3.setAttribute('data-start-pos', '60');
div3.appendChild(span3);

const res3 = calculateSourceOffset(div3, 0);
assert(res3 === 60, `Element start offset. Got ${res3}, want 60`);

// Test 4: Fallback to ancestor
// <div data-start-pos="200"><span>Text</span></div>
// Select "Text" offset 0.
// Span has no data-pos. Parent has 200.
// Should return 200.
const div4 = new MockNode(Node.ELEMENT_NODE);
div4.setAttribute('data-start-pos', '200');
const span4 = new MockNode(Node.ELEMENT_NODE); // No pos
div4.appendChild(span4);
const text4 = new MockNode(Node.TEXT_NODE, "Text");
span4.appendChild(text4);

const res4 = calculateSourceOffset(text4, 0);
assert(res4 === 200, `Ancestor fallback. Got ${res4}, want 200`);

// Test 5: End pos
// <div data-end-pos="300"></div>
// Offset 0 (no children).
// Logic: if offset < childNodes.length ... else use end-pos.
// childNodes is empty. 0 is not < 0.
// So returns data-end-pos.
const div5 = new MockNode(Node.ELEMENT_NODE);
div5.setAttribute('data-end-pos', '300');
const res5 = calculateSourceOffset(div5, 0);
assert(res5 === 300, `End pos. Got ${res5}, want 300`);

console.log("All JS tests passed.");
