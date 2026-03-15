<template>
  <div class="discussion-page">
    <div class="discussion-title">社区讨论</div>

    <div class="section-tabs">
      <button
        v-for="s in sections"
        :key="s.key"
        class="tab-btn"
        :class="{ on: currentSection === s.key }"
        @click="switchSection(s.key)"
      >{{ s.name }}</button>
    </div>

    <div class="toolbar">
      <button class="primary" @click="createOpen = true">发帖</button>
    </div>

    <div class="post-list">
      <div class="post-item" v-for="p in currentPosts" :key="p.id">
        <router-link class="post-title" :to="`/discussion/${currentSection}/${p.id}`">{{ p.title }}</router-link>
        <div class="post-meta">{{ formatTime(p.createdAt) }} · 回复 {{ p.replies.length }}</div>
      </div>
      <div class="empty" v-if="currentPosts.length===0">本分区还没有帖子，来发第一条吧。</div>
    </div>

    <div v-if="createOpen" class="modal-mask" @click.self="createOpen = false">
      <div class="modal">
        <div class="modal-title">发布新帖</div>
        <input v-model.trim="newTitle" class="input" placeholder="标题" />
        <textarea v-model.trim="newContent" class="textarea" placeholder="内容"></textarea>
        <div class="modal-actions">
          <button class="primary" @click="submitPost">发布</button>
          <button class="ghost" @click="createOpen = false">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class DiscussionPage extends Vue {
  sections: any[] = [
    { key: 'use', name: '使用交流' },
    { key: 'bug', name: '报错反馈' },
    { key: 'teach', name: '教学分享' },
    { key: 'gallery', name: '画册' },
  ]
  currentSection: string = 'use'
  postsBySection: any = { use: [], bug: [], teach: [], gallery: [] }

  createOpen: boolean = false
  newTitle: string = ''
  newContent: string = ''

  get currentPosts() {
    return this.postsBySection[this.currentSection] || []
  }

  created() {
    this.loadPosts()
  }

  loadPosts() {
    try {
      const raw = localStorage.getItem('wm_discussion_posts') || '{}'
      const obj = JSON.parse(raw)
      this.postsBySection = { use: [], bug: [], teach: [], gallery: [], ...obj }
    } catch (e) {
      this.postsBySection = { use: [], bug: [], teach: [], gallery: [] }
    }
  }

  savePosts() {
    localStorage.setItem('wm_discussion_posts', JSON.stringify(this.postsBySection))
  }

  switchSection(key: string) {
    this.currentSection = key
  }

  submitPost() {
    if (!this.newTitle || !this.newContent) return
    const p = {
      id: 'p_' + Date.now() + '_' + Math.random().toString(36).slice(2, 6),
      title: this.newTitle,
      content: this.newContent,
      createdAt: Date.now(),
      replies: [] as any[],
    }
    this.postsBySection[this.currentSection].unshift(p)
    this.savePosts()
    this.createOpen = false
    this.newTitle = ''
    this.newContent = ''
  }

  formatTime(ts: number) {
    const d = new Date(ts)
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n)
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}
</script>

<style scoped lang="stylus">
.discussion-page
  max-width 980px
  margin 18px auto
  background rgba(255,255,255,.82)
  border 1px solid rgba(0,0,0,.1)
  border-radius 14px
  padding 14px

.discussion-title
  font-size 22px
  font-weight 800
  margin-bottom 10px

.section-tabs
  display flex
  gap 8px
  flex-wrap wrap

.tab-btn
  border 1px solid rgba(60,120,255,.35)
  background rgba(220,235,255,.75)
  color rgba(30,90,210,1)
  border-radius 10px
  padding 6px 12px
  cursor pointer

.tab-btn.on
  background rgba(60,120,255,.85)
  color #fff

.toolbar
  margin 12px 0

.primary
  border 1px solid rgba(60,120,255,.35)
  background rgba(220,235,255,.85)
  color rgba(30,90,210,1)
  border-radius 10px
  padding 6px 12px
  cursor pointer

.ghost
  border 1px solid rgba(0,0,0,.2)
  background rgba(255,255,255,.85)
  border-radius 10px
  padding 6px 12px
  cursor pointer

.post-item
  border 1px solid rgba(0,0,0,.12)
  border-radius 10px
  padding 10px
  margin-bottom 8px
  background #fff

.post-title
  color #2f6bff
  text-decoration none
  font-weight 800

.post-meta
  margin-top 4px
  font-size 12px
  opacity .7

.empty
  opacity .7

.modal-mask
  position fixed
  inset 0
  background rgba(0,0,0,.35)
  display flex
  align-items center
  justify-content center

.modal
  width 520px
  max-width 92vw
  background #fff
  border-radius 12px
  padding 14px

.modal-title
  font-weight 800
  margin-bottom 8px

.input, .textarea
  width 100%
  border 1px solid rgba(0,0,0,.2)
  border-radius 10px
  padding 8px
  margin-bottom 10px

.textarea
  min-height 120px
  resize vertical

.modal-actions
  display flex
  justify-content flex-end
  gap 8px
</style>
