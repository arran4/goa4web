const assert = require('assert');

// Mock DOM and environment for pasteimg.js
global.document = {
    querySelectorAll: () => [],
    querySelector: () => null,
    createElement: () => ({}),
    addEventListener: () => {}
};
global.window = {
    addEventListener: () => {}
};
global.FileReader = class {};
global.FormData = class {};
global.XMLHttpRequest = class {
    open() {}
    send() {}
};

// We need to extract the logic to test it.
const fs = require('fs');
const code = fs.readFileSync('./core/templates/assets/pasteimg.js', 'utf8');

// The logic is wrapped in an IIFE. Let's just run it and grab the functions we need by injecting them into global.
const scriptToRun = code.replace(
    /function isEscaped\(text, index\) {/,
    'global.isEscaped = isEscaped; global.isInsideCodeBlock = isInsideCodeBlock; global.escapeCodeBlockContent = escapeCodeBlockContent; global.handleTextPaste = handleTextPaste; function isEscaped(text, index) {'
);

eval(scriptToRun);

console.log("Running Code Block Detection Tests (Parity)...");
assert.strictEqual(global.isInsideCodeBlock("[code \n"), true, "[code is detected");
assert.strictEqual(global.isInsideCodeBlock("\\[code \n"), false, "escaped \\[code is NOT detected");
assert.strictEqual(global.isInsideCodeBlock("\\\\[code \n"), true, "double escaped \\\\[code is detected");
assert.strictEqual(global.isInsideCodeBlock("\\\\\\[code \n"), false, "triple escaped \\\\\\[code is NOT detected");
assert.strictEqual(global.isInsideCodeBlock("[code]"), true, "[code] leaves it open");
assert.strictEqual(global.isInsideCodeBlock("[code] "), true, "[code] space leaves it open");
assert.strictEqual(global.isInsideCodeBlock("[code] text ]"), false, "closing bracket terminates block");
assert.strictEqual(global.isInsideCodeBlock("[code] text \\]"), true, "escaped bracket doesn't terminate");
assert.strictEqual(global.isInsideCodeBlock("[code] text \\\\]"), false, "double escaped bracket terminates");
assert.strictEqual(global.isInsideCodeBlock("[code] text \\\\\\]"), true, "triple escaped bracket doesn't terminate");

console.log("Running Escaping Tests (Parity)...");
assert.strictEqual(global.escapeCodeBlockContent("text ] text"), "text \\] text", "escapes unescaped ]");
assert.strictEqual(global.escapeCodeBlockContent("text \\] text"), "text \\] text", "preserves escaped \\]");
assert.strictEqual(global.escapeCodeBlockContent("text \\\\] text"), "text \\\\\\] text", "escapes double escaped \\\\]");
assert.strictEqual(global.escapeCodeBlockContent("text \\\\\\] text"), "text \\\\\\] text", "preserves triple escaped \\\\\\]");
assert.strictEqual(global.escapeCodeBlockContent("text]"), "text\\]", "escapes end ]");
assert.strictEqual(global.escapeCodeBlockContent("]text"), "\\]text", "escapes start ]");

console.log("Running Text Paste Mock Tests...");
let dispatched = false;
let prevented = false;
let mockTextarea = {
    value: "[code \n",
    selectionStart: 7,
    selectionEnd: 7,
    setRangeText: function(replacement, start, end, selectionMode) {
        this.value = this.value.substring(0, start) + replacement + this.value.substring(end);
        this.selectionStart = start + replacement.length;
        this.selectionEnd = start + replacement.length;
    },
    dispatchEvent: function(e) {
        if (e.type === 'input') {
            dispatched = true;
        }
    }
};

let mockEvent = {
    target: mockTextarea,
    clipboardData: {
        getData: (type) => "hello ] world \\\\] test \\]"
    },
    preventDefault: () => { prevented = true; }
};

let handled = global.handleTextPaste(mockEvent);
assert.strictEqual(handled, true, "handled should be true inside code block");
assert.strictEqual(prevented, true, "default should be prevented");
assert.strictEqual(dispatched, true, "input event should be dispatched");
assert.strictEqual(mockTextarea.value, "[code \nhello \\] world \\\\\\] test \\]", "pasted text should be properly escaped and inserted");

console.log("All parity and paste tests passed!");
