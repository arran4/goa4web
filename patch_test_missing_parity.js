const fs = require('fs');

let testCode = fs.readFileSync('tests/js/site.test.js', 'utf8');

const testAdds = `
    checkParity(global.isInsideCodeBlock("[CODE \\n"), true, "[CODE is detected");
    checkParity(global.isInsideCodeBlock("[CoDe]"), true, "[CoDe] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\"]"), true, "[codein \\"go\\"] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\" "), true, "[codein \\"go\\" space leaves it open");
    checkParity(global.isInsideCodeBlock("[codein go]"), true, "[codein go] leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"g\\\\]o\\"]"), true, "[codein \\"g\\\\]o\\"] escapes inside quotes and leaves it open");
    checkParity(global.isInsideCodeBlock("[codein \\"go\\" ]"), false, "[codein \\"go\\" ] closes because ] terminates");
    checkParity(global.isInsideCodeBlock("[codein \\"g\\\\]"), false, "Still in codein arg");
`;

testCode = testCode.replace(
    /checkParity\(global\.escapeCodeBlockContent\("text \] text"\), "text \\\\\] text", "escapes unescaped \]"\);/,
    `checkParity(global.escapeCodeBlockContent("text ] text"), "text \\\\] text", "escapes unescaped ]");
${testAdds}`
);

fs.writeFileSync('tests/js/site.test.js', testCode);
