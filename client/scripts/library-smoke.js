// Run with: node scripts/library-smoke.js
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const vm = require('vm');
const ts = require('typescript');
const compiler = require('vue-template-compiler');
const Vue = require('vue');
const source = compiler.parseComponent(fs.readFileSync(path.join(__dirname, '../src/components/ConversationLibrary.vue'), 'utf8'));
const template = compiler.compile(source.template.content);
assert.deepStrictEqual(template.errors, []);
let liked = false;
let comments = [];
let publication;
const entry = () => ({ id: 1, title: 'Story', author: 'reader', messages: publication.messages, likes: liked ? 1 : 0, liked, comments: comments.length });
const axios = async config => {
  assert.strictEqual(config.headers.Authorization, 'Bearer test-token');
  let data;
  if (config.method === 'POST' && config.url === '/api/library') { publication = config.data; data = { id: 1 }; }
  else if (config.url.includes('/like')) { liked = config.data.liked; data = entry(); }
  else if (config.url.includes('/comments') && config.method === 'POST') { comments.push({ id: 1, content: config.data.content }); data = { ok: true }; }
  else if (config.url.includes('/comments')) data = comments;
  else if (config.url.includes('?offset=')) data = publication ? [entry()] : [];
  else data = entry();
  return { data: { code: 200, data } };
};
const moduleObject = { exports: {} };
vm.runInNewContext(ts.transpileModule(source.script.content, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2018, experimentalDecorators: true, esModuleInterop: true } }).outputText, {
  module: moduleObject, exports: moduleObject.exports,
  require: name => name === 'axios' ? axios : require(name),
  document: { activeElement: { focus() {} } }, console, setTimeout, clearTimeout
});
async function main() {
  const Library = moduleObject.exports.default;
  const messages = [{ name: 'You', content: '<script>literal</script>\nhello' }];
  const library = new Library({ propsData: { messages, suggestedTitle: 'Story', token: 'test-token' } });
  assert.strictEqual(library.expanded, true);
  library.preview();
  messages[0].content = 'private continuation';
  assert.strictEqual(library.snapshot[0].content, '<script>literal</script>\nhello');
  await library.publish();
  assert.strictEqual(library.dialog, 'read');
  assert.strictEqual(library.detail.id, 1);
  assert.strictEqual(publication.messages[0].content, '<script>literal</script>\nhello');
  await library.like(); assert.strictEqual(library.detail.likes, 1);
  await library.like(); assert.strictEqual(library.detail.likes, 0);
  library.comment = 'A comment'; await library.postComment();
  assert.strictEqual(library.comment, ''); assert.strictEqual(library.comments.length, 1);
  library.close(); assert.strictEqual(library.dialog, '');
  const guest = new Library({ propsData: { messages } });
  let loginRequested = false; guest.$on('login', () => { loginRequested = true; });
  guest.preview(); assert(loginRequested); assert.strictEqual(guest.dialog, '');
  await Vue.nextTick(); library.$destroy(); guest.$destroy();
  console.log('PASS: template, default expansion, publication preview snapshot, publish/read, like/unlike, comment, close, guest login');
}
main().catch(error => { console.error(error); process.exitCode = 1; });
