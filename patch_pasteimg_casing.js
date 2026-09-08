const fs = require('fs');
let code = fs.readFileSync('core/templates/assets/pasteimg.js', 'utf8');

const oldLogic = `
    function isInsideCodeBlock(textBeforeCaret) {
        let i = 0;
        let inCodeBlock = false;

        while (i < textBeforeCaret.length) {
            if (!inCodeBlock && textBeforeCaret.substring(i).startsWith('[code') && !isEscaped(textBeforeCaret, i)) {
                let nextChar = textBeforeCaret[i + 5];
                if (nextChar === ' ' || nextChar === '\\n' || nextChar === '\\r' || nextChar === ']' || !nextChar) {
                    inCodeBlock = true;
                    i += 5;
                    if (textBeforeCaret[i] === ']') {
                        i++;
                    }
                    continue;
                }
            }
            if (inCodeBlock && textBeforeCaret[i] === ']' && !isEscaped(textBeforeCaret, i)) {
                inCodeBlock = false;
            }
            i++;
        }
        return inCodeBlock;
    }
`;

const newLogic = `
    function isInsideCodeBlock(textBeforeCaret) {
        let i = 0;
        let inCodeBlock = false;

        const lowerText = textBeforeCaret.toLowerCase();

        while (i < lowerText.length) {
            if (!inCodeBlock && !isEscaped(lowerText, i)) {
                if (lowerText.substring(i).startsWith('[codein')) {
                    let nextChar = lowerText[i + 7];
                    if (nextChar === ' ' || nextChar === '\\n' || nextChar === '\\r' || nextChar === ']' || nextChar === '=' || !nextChar) {
                        // For codein, it requires an argument. We need to skip the argument block.
                        // The argument ends at the first unescaped space, ], \\n, \\r, =
                        let j = i + 7;

                        // We skip whitespace
                        while (j < lowerText.length && (lowerText[j] === ' ' || lowerText[j] === '\\n' || lowerText[j] === '\\r')) {
                            j++;
                        }

                        // Wait, a4code uses GetNextArg which checks if it starts with ".
                        if (j < lowerText.length && lowerText[j] === '"') {
                            j++;
                            while (j < lowerText.length) {
                                if (lowerText[j] === '"' && !isEscaped(lowerText, j)) {
                                    j++;
                                    break;
                                }
                                j++;
                            }
                        } else {
                            // reads until space, ], [, \\n, \\r, =
                            while (j < lowerText.length) {
                                if ((lowerText[j] === ' ' || lowerText[j] === ']' || lowerText[j] === '[' || lowerText[j] === '\\n' || lowerText[j] === '\\r' || lowerText[j] === '=') && !isEscaped(lowerText, j)) {
                                    break;
                                }
                                j++;
                            }
                        }

                        // Skip prefix space after arg
                        while (j < lowerText.length && (lowerText[j] === ' ' || lowerText[j] === '\\n' || lowerText[j] === '\\r' || lowerText[j] === '=')) {
                            j++;
                        }

                        // Consume code block content bytes until terminator
                        inCodeBlock = true;
                        i = j;

                        // If it ends with ] immediately, that's just the terminator in some legacy syntax or if it's empty
                        if (i < lowerText.length && lowerText[i] === ']') {
                            i++;
                        }
                        continue;
                    }
                } else if (lowerText.substring(i).startsWith('[code')) {
                    let nextChar = lowerText[i + 5];
                    if (nextChar === ' ' || nextChar === '\\n' || nextChar === '\\r' || nextChar === ']' || nextChar === '=' || !nextChar) {
                        inCodeBlock = true;
                        i += 5;

                        // skipArgPrefix logic in a4code parser for code
                        while (i < lowerText.length && (lowerText[i] === ' ' || lowerText[i] === '\\n' || lowerText[i] === '\\r' || lowerText[i] === '=')) {
                            i++;
                        }

                        if (i < lowerText.length && lowerText[i] === ']') {
                            i++;
                        }
                        continue;
                    }
                }
            }
            if (inCodeBlock && lowerText[i] === ']' && !isEscaped(lowerText, i)) {
                inCodeBlock = false;
            }
            i++;
        }
        return inCodeBlock;
    }
`;

code = code.replace(oldLogic, newLogic);
fs.writeFileSync('core/templates/assets/pasteimg.js', code);

// Now patch the tests
let testCode = fs.readFileSync('tests/js/site.test.js', 'utf8');

const testAdds = `
    checkParity(global.isInsideCodeBlock("[CODE \\n"), true, "[CODE is detected");
    checkParity(global.isInsideCodeBlock("[CoDe]"), true, "[CoDe] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\"]"), true, "[codein \\"go\\"] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\" "), true, "[codein \\"go\\" space leaves it open");
    checkParity(global.isInsideCodeBlock("[codein go]"), true, "[codein go] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"g\\\\]o\\"]"), true, "[codein \\"g\\\\]o\\"] escapes inside quotes and leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\" ]"), false, "[codein \\"go\\" ] closes because ] terminates");
    checkParity(global.isInsideCodeBlock("[codein \\"g\\\\]"), false, "Still in codein arg"); // Wait, if we are at the caret right after g\\\\] in [codein "g\\]
    checkParity(global.isInsideCodeBlock("[codein \\"g\\\\]"), false, "caret in arg is not code content");
`;

testCode = testCode.replace(
    /checkParity\(global\.isInsideCodeBlock\("\\[code\] text \\\\\\\\\\\\\\\\\\]"\), false, "double escaped bracket terminates"\);/,
    `checkParity(global.isInsideCodeBlock("[code] text \\\\\\\\\\]"), false, "double escaped bracket terminates");
${testAdds}`
);

fs.writeFileSync('tests/js/site.test.js', testCode);
