const { describe, it } = require('node:test');
const assert = require('node:assert');
const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

const html = fs.readFileSync(path.resolve(__dirname, './index.html'), 'utf8');

describe('Jabra UI Frontend', () => {
  it('Deve renderizar os elementos da Mini View por padrão', () => {
    const dom = new JSDOM(html);
    const { document } = dom.window;
    
    assert.strictEqual(document.body.className, 'mini-view');
    assert.ok(document.getElementById('battery-level'));
    assert.ok(document.getElementById('custom-id'));
    assert.ok(document.getElementById('clock'));
  });

  it('Deve ocultar o histórico na Mini View', () => {
    const dom = new JSDOM(html);
    const { document } = dom.window;
    const fullViewOnly = document.getElementById('full-view-only');
    
    assert.strictEqual(fullViewOnly.style.display, 'none');
  });
});
