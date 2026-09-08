<template>
  <div class="library-root">
    <aside v-show="expanded" id="public-library" class="library-panel" aria-label="公开对话图书馆">
      <header><strong>图书馆</strong><button @click="expanded = false" aria-label="收起图书馆">‹ 收起</button></header>
      <p class="muted">读一段对话，留一点共鸣。</p>
      <button class="publish" @click="preview">记录进图书馆</button>
      <div class="library-tools"><span>所有公开对话</span><button @click="loadEntries(true)" :disabled="loading">刷新</button></div>
      <p v-if="listError" role="alert">{{ listError }}</p>
      <div class="library-list">
        <button v-for="entry in entries" :key="entry.id" class="library-book" @click="openEntry(entry.id)">
          <strong>{{ entry.title }}</strong><span>{{ entry.author }} · {{ date(entry.created_at) }}</span>
          <span>♡ {{ entry.likes }} · 留言 {{ entry.comments }}</span>
        </button>
        <p v-if="!loading && !entries.length && !listError" class="muted">还没有公开记录，来收藏第一段故事吧。</p>
        <p v-if="loading" role="status">正在打开书架…</p>
        <button v-if="hasMore && entries.length" @click="loadEntries(false)" :disabled="loading">加载更多</button>
      </div>
    </aside>
    <button v-if="!expanded" class="library-tab" aria-controls="public-library" :aria-expanded="expanded.toString()" @click="expanded = true">图<br>书<br>馆 ›</button>

    <div v-if="dialog" class="library-mask" @click.self="close" @keydown.esc.stop="close" @keydown.tab="trapFocus">
      <section ref="dialog" class="library-dialog" role="dialog" aria-modal="true" aria-labelledby="library-title" tabindex="-1">
        <header><h2 id="library-title">{{ dialog === 'publish' ? '记录进图书馆' : (detail ? detail.title : '正在打开记录…') }}</h2><button @click="close" :disabled="publishing">关闭</button></header>
        <div class="library-body">
          <template v-if="dialog === 'publish'">
            <p class="public-notice">发布后所有人（包括未登录访客）都能看到以下完整对话和你的用户名。只保存本次快照，之后的聊天不会自动公开。</p>
            <label for="library-publish-title">记录标题</label>
            <input id="library-publish-title" v-model="title" maxlength="80" placeholder="给这段对话起个名字" :disabled="publishing" />
            <p class="muted">共 {{ snapshot.length }} 条消息，请检查后再公开。</p>
            <article v-for="(message, index) in snapshot" :key="index" class="library-message"><b>{{ message.name }}</b><p>{{ message.content }}</p></article>
          </template>
          <template v-else-if="detail">
            <p class="muted">{{ detail.author }} · {{ date(detail.created_at) }} · {{ detail.messages.length }} 条消息</p>
            <article v-for="(message, index) in detail.messages" :key="index" class="library-message"><b>{{ message.name }}</b><p>{{ message.content }}</p></article>
            <button class="like" :aria-pressed="detail.liked.toString()" :disabled="liking" @click="like">{{ detail.liked ? '♥ 已点赞' : '♡ 点赞' }} {{ detail.likes }}</button>
            <h3>留言 · {{ detail.comments }}</h3>
            <form @submit.prevent="postComment">
              <label for="library-comment">写下你的留言</label>
              <textarea id="library-comment" v-model="comment" maxlength="1000" rows="3" :disabled="commenting" placeholder="聊聊你读完的感受…"></textarea>
              <button :disabled="commenting || !comment.trim()">{{ commenting ? '发布中…' : '发表留言' }}</button>
            </form>
            <p v-if="commentsLoading" role="status">正在加载留言…</p>
            <p v-if="!commentsLoading && !comments.length" class="muted">还没有留言。</p>
            <article v-for="item in comments" :key="item.id" class="library-message"><b>{{ item.author }}</b> <small>{{ date(item.created_at) }}</small><p>{{ item.content }}</p></article>
            <button v-if="commentsMore" @click="loadComments(false)" :disabled="commentsLoading">更多留言</button>
          </template>
          <p v-if="error" class="library-error" role="alert">{{ error }}</p>
          <button v-if="dialog === 'read' && !detail && error" @click="openEntry(selectedId)">重新加载</button>
        </div>
        <footer v-if="dialog === 'publish'"><button class="publish" @click="publish" :disabled="publishing || !title.trim() || !snapshot.length">{{ publishing ? '正在公开…' : '确认公开这段对话' }}</button></footer>
      </section>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Prop, Watch, Vue } from 'vue-property-decorator';
import axios from 'axios';

@Component
export default class ConversationLibrary extends Vue {
  @Prop({ default: () => [] }) readonly messages!: any[];
  @Prop({ default: '' }) readonly suggestedTitle!: string;
  @Prop({ default: '' }) readonly token!: string;
  expanded = true;
  entries: any[] = [];
  loading = false;
  hasMore = true;
  listError = '';
  dialog = '';
  detail: any = null;
  selectedId = 0;
  snapshot: any[] = [];
  title = '';
  error = '';
  publishing = false;
  liking = false;
  commenting = false;
  commentsLoading = false;
  comments: any[] = [];
  commentsMore = false;
  comment = '';
  private requestVersion = 0;
  private returnFocus: HTMLElement | null = null;

  mounted() { this.loadEntries(true); }
  beforeDestroy() { this.requestVersion++; }
  @Watch('token') onTokenChanged() {
    this.loadEntries(true);
    if (this.dialog === 'read' && this.selectedId) this.openEntry(this.selectedId);
  }
  private async request(path: string, method: string = 'GET', data?: any) {
    try {
      const res = await axios({ url: '/api/library' + path, method, data, timeout: 20000,
        headers: this.token ? { Authorization: 'Bearer ' + this.token } : {} });
      if (res.data.code !== 200) throw new Error(res.data.data.error || '请求失败，请重试');
      return res.data.data;
    } catch (e) {
      const err: any = e;
      if (err.response && err.response.status === 401) this.$emit('login');
      throw new Error((err.response && err.response.data && err.response.data.data && err.response.data.data.error) || err.message || '网络连接失败，请重试');
    }
  }
  async loadEntries(reset: boolean) {
    if (this.loading) return;
    this.loading = true; this.listError = '';
    try {
      const rows = await this.request('?offset=' + (reset ? 0 : this.entries.length));
      this.entries = reset ? rows : this.entries.concat(rows);
      this.hasMore = rows.length === 30;
    } catch (e) { this.listError = (e as Error).message; }
    finally { this.loading = false; }
  }
  private showDialog(kind: string) {
    if (!this.dialog) this.returnFocus = document.activeElement as HTMLElement;
    this.dialog = kind;
    this.$nextTick(() => { const el = this.$refs.dialog as HTMLElement; if (el) el.focus(); });
  }
  preview() {
    if (!this.token) { this.$emit('login'); return; }
    this.snapshot = this.messages.map(m => ({ name: String(m.name), content: String(m.content) }));
    this.title = this.suggestedTitle.slice(0, 80) || '与魔理沙的对话';
    this.error = this.snapshot.length ? '' : '当前还没有可以公开的对话。';
    this.showDialog('publish');
  }
  async publish() {
    if (this.publishing || !this.title.trim() || !this.snapshot.length) return;
    this.publishing = true; this.error = '';
    try {
      const result = await this.request('', 'POST', { title: this.title, messages: this.snapshot });
      this.publishing = false;
      this.loadEntries(true);
      await this.openEntry(result.id);
    } catch (e) { this.error = (e as Error).message; }
    finally { this.publishing = false; }
  }
  async openEntry(id: number) {
    const version = ++this.requestVersion;
    this.selectedId = id; this.detail = null; this.error = ''; this.comments = []; this.comment = ''; this.commentsMore = false;
    this.showDialog('read');
    try {
      const item = await this.request('/' + id);
      if (version !== this.requestVersion) return;
      this.detail = item;
      await this.loadComments(true);
    } catch (e) { if (version === this.requestVersion) this.error = (e as Error).message; }
  }
  async loadComments(reset: boolean) {
    const version = this.requestVersion;
    this.commentsLoading = true;
    try {
      const rows = await this.request('/' + this.selectedId + '/comments?offset=' + (reset ? 0 : this.comments.length));
      if (version !== this.requestVersion) return;
      this.comments = reset ? rows : this.comments.concat(rows); this.commentsMore = rows.length === 30;
    } catch (e) { if (version === this.requestVersion) this.error = (e as Error).message; }
    finally { if (version === this.requestVersion) this.commentsLoading = false; }
  }
  async like() {
    if (!this.token) { this.$emit('login'); return; }
    if (this.liking || !this.detail) return;
    const version = this.requestVersion;
    this.liking = true; this.error = '';
    try {
      const item = await this.request('/' + this.selectedId + '/like', 'PUT', { liked: !this.detail.liked });
      if (version === this.requestVersion) this.detail = item;
      this.loadEntries(true);
    } catch (e) { if (version === this.requestVersion) this.error = (e as Error).message; }
    finally { this.liking = false; }
  }
  async postComment() {
    if (!this.token) { this.$emit('login'); return; }
    if (this.commenting || !this.comment.trim()) return;
    const version = this.requestVersion;
    this.commenting = true; this.error = '';
    try {
      await this.request('/' + this.selectedId + '/comments', 'POST', { content: this.comment });
      if (version === this.requestVersion) { this.comment = ''; this.detail.comments++; await this.loadComments(true); }
      this.loadEntries(true);
    } catch (e) { if (version === this.requestVersion) this.error = (e as Error).message; }
    finally { this.commenting = false; }
  }
  close() {
    if (this.publishing) return;
    this.requestVersion++; this.dialog = '';
    if (this.returnFocus) this.returnFocus.focus();
  }
  trapFocus(event: KeyboardEvent) {
    const el = this.$refs.dialog as HTMLElement;
    const nodes = el.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), textarea:not(:disabled)');
    const first = nodes[0], last = nodes[nodes.length - 1];
    if (event.shiftKey && (document.activeElement === first || document.activeElement === el)) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && (document.activeElement === last || document.activeElement === el)) { event.preventDefault(); first.focus(); }
  }
  date(value: number) { return new Date(value * 1000).toLocaleString(); }
}
</script>

<style scoped>
.library-root { color: #293444; font: 14px/1.6 sans-serif; }
.library-panel { position: fixed; left: 0; top: 86px; bottom: 20px; width: 252px; z-index: 10010; padding: 16px; background: rgba(250,248,240,.97); border: 1px solid #d9d5c9; border-radius: 0 16px 16px 0; box-shadow: 4px 8px 28px #0002; display: flex; flex-direction: column; gap: 10px; }
header, .library-tools { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
header strong { font-size: 20px; }
button { cursor: pointer; border: 1px solid #c6cbd2; border-radius: 8px; padding: 7px 10px; background: #fff; color: #293444; font: inherit; }
button:disabled { opacity: .55; cursor: not-allowed; }
button:hover:not(:disabled) { background: #eaf0f8; }
button:focus-visible, input:focus-visible, textarea:focus-visible { outline: 3px solid #4774ab; outline-offset: 2px; }
.publish { background: #355b83; color: #fff; }
.publish:hover:not(:disabled) { background: #28496e; }
.muted, small { color: #647080; font-size: 12px; }
p { margin: 0; }
.library-list { overflow-y: auto; min-height: 0; }
.library-book { width: 100%; display: flex; flex-direction: column; align-items: flex-start; text-align: left; margin-bottom: 10px; padding: 12px; overflow-wrap: anywhere; }
.library-book span { font-size: 12px; color: #647080; }
.library-tab { position: fixed; top: 40%; left: 0; z-index: 10010; border-radius: 0 10px 10px 0; background: #faf8f0; box-shadow: 2px 4px 16px #0002; }
.library-mask { position: fixed; inset: 0; z-index: 12500; background: #14203099; display: flex; align-items: center; justify-content: center; padding: 16px; }
.library-dialog { width: 760px; max-width: 100%; max-height: 88vh; max-height: 88dvh; background: #faf9f5; border-radius: 18px; display: flex; flex-direction: column; box-shadow: 0 20px 70px #0005; }
.library-dialog header, footer { padding: 16px 20px; flex-shrink: 0; border-bottom: 1px solid #deded7; }
h2 { font-size: 18px; margin: 0; overflow-wrap: anywhere; }
.library-body { overflow-y: auto; padding: 20px; overscroll-behavior: contain; }
.library-message { margin: 12px 0; padding: 12px 14px; background: #fff; border: 1px solid #e2e3df; border-radius: 10px; overflow-wrap: anywhere; }
.library-message p { white-space: pre-wrap; margin-top: 6px; }
.public-notice { padding: 12px; background: #fff1cf; border-radius: 8px; margin-bottom: 14px; }
input, textarea { width: 100%; box-sizing: border-box; padding: 10px; border: 1px solid #bcc5d0; border-radius: 8px; background: #fff; color: #293444; font: inherit; margin: 6px 0 10px; }
textarea { resize: vertical; }
.library-error { color: #ac3030; margin-top: 12px; }
.like[aria-pressed=true] { background: #ffe6e9; color: #a13151; }
@media(max-width:900px) { .library-panel { top: 72px; bottom: auto; max-height: 48vh; width: min(252px, 80vw); } .library-mask { padding: 8px; } .library-dialog { max-height: 90vh; max-height: 90dvh; } .library-body { padding: 12px; } input, textarea { font-size: 16px; } }
</style>
