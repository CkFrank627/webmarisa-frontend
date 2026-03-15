<template>
  <div class="post-page" v-if="post">
    <div class="back-line">
      <router-link :to="'/discussion'">← 返回讨论页</router-link>
    </div>
    <h1 class="title">{{ post.title }}</h1>
    <div class="meta">{{ formatTime(post.createdAt) }}</div>
    <div class="content">{{ post.content }}</div>

    <div class="reply-box">
      <div class="reply-title">留言回复</div>
      <textarea v-model.trim="newReply" class="textarea" placeholder="写下你的回复..."></textarea>
      <button class="primary" @click="submitReply">发送回复</button>
    </div>

    <div class="reply-list">
      <div class="reply-item" v-for="r in post.replies" :key="r.id">
        <div class="reply-meta">{{ formatTime(r.createdAt) }}</div>
        <div class="reply-content">{{ r.content }}</div>
      </div>
      <div class="empty" v-if="post.replies.length===0">还没有回复</div>
    </div>
  </div>
  <div v-else class="post-page">帖子不存在或已被删除。</div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class DiscussionPostPage extends Vue {
  post: any = null
  section: string = ''
  postId: string = ''
  newReply: string = ''

  created() {
    this.section = String(this.$route.params.section || '')
    this.postId = String(this.$route.params.postId || '')
    this.loadPost()
  }

  loadPost() {
    try {
      const raw = localStorage.getItem('wm_discussion_posts') || '{}'
      const obj = JSON.parse(raw)
      const list = obj[this.section] || []
      this.post = list.find((p: any) => p.id === this.postId) || null
      if (this.post && !Array.isArray(this.post.replies)) this.post.replies = []
    } catch (e) {
      this.post = null
    }
  }

  savePost(updated: any) {
    const raw = localStorage.getItem('wm_discussion_posts') || '{}'
    const obj = JSON.parse(raw)
    const list = obj[this.section] || []
    obj[this.section] = list.map((p: any) => p.id === this.postId ? updated : p)
    localStorage.setItem('wm_discussion_posts', JSON.stringify(obj))
  }

  submitReply() {
    if (!this.post || !this.newReply) return
    const r = {
      id: 'r_' + Date.now() + '_' + Math.random().toString(36).slice(2, 6),
      content: this.newReply,
      createdAt: Date.now(),
    }
    this.post.replies.push(r)
    this.savePost(this.post)
    this.newReply = ''
  }

  formatTime(ts: number) {
    const d = new Date(ts)
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n)
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}
</script>

<style scoped lang="stylus">
.post-page
  max-width 980px
  margin 18px auto
  background rgba(255,255,255,.84)
  border 1px solid rgba(0,0,0,.1)
  border-radius 14px
  padding 14px

.back-line
  margin-bottom 8px

.title
  margin 0 0 6px

.meta
  font-size 12px
  opacity .7

.content
  margin 12px 0
  line-height 1.7
  white-space pre-wrap

.reply-box
  border-top 1px solid rgba(0,0,0,.1)
  padding-top 12px

.reply-title
  font-weight 800
  margin-bottom 8px

.textarea
  width 100%
  min-height 90px
  border 1px solid rgba(0,0,0,.2)
  border-radius 10px
  padding 8px
  margin-bottom 8px
  resize vertical

.primary
  border 1px solid rgba(60,120,255,.35)
  background rgba(220,235,255,.85)
  color rgba(30,90,210,1)
  border-radius 10px
  padding 6px 12px
  cursor pointer

.reply-list
  margin-top 14px

.reply-item
  border 1px solid rgba(0,0,0,.12)
  border-radius 10px
  padding 8px
  margin-bottom 8px
  background #fff

.reply-meta
  font-size 12px
  opacity .7
  margin-bottom 4px

.reply-content
  white-space pre-wrap

.empty
  opacity .7
</style>
