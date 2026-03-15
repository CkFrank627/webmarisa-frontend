<template>
  <div id="app">
    <header class="site-header" :class="{ scrolled: headerScrolled }">
      <div class="site-header-inner">
        <router-link class="brand" to="/" title="返回主页面">
          <div class="brand-avatar">魔</div>
          <div class="brand-text">
            <div class="brand-title">白丝魔理沙</div>
            <div class="brand-sub">{{ headerMotto }}</div>
          </div>
        </router-link>

        <nav class="nav-group">
          <router-link class="nav-btn" to="/">主页面</router-link>
          <router-link class="nav-btn" to="/discussion">讨论页</router-link>
          <button class="nav-btn" @click="qqModalOpen = true">加入群聊</button>

          <div class="support-wrap">
            <button
              class="nav-btn support"
              :class="{ done: supported }"
              @click="supportOwner"
              @mouseenter="showSupportTip = true"
              @mouseleave="showSupportTip = false"
              :title="supportTipText"
            >
              {{ supported ? '已支持站长' : '支持站长' }}
            </button>
            <div class="support-tip" v-if="showSupportTip && supported">已支持人数：{{ supportCount }}</div>
          </div>
        </nav>

        <div class="header-tools">
          <button class="tool-btn" @click="nextMotto" title="换一句今日寄语">✨ 今日寄语</button>
          <div class="clock">{{ currentTime }}</div>
        </div>
      </div>
    </header>

    <router-view/>

    <div v-if="qqModalOpen" class="modal-mask" @click.self="qqModalOpen = false">
      <div class="modal">
        <div class="modal-title">加入群聊</div>
        <div class="modal-content">QQ号：<b>874128517</b></div>
        <div class="modal-actions">
          <button class="nav-btn dark" @click="qqModalOpen = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class App extends Vue {
  qqModalOpen: boolean = false
  supportCount: number = 0
  supported: boolean = false
  showSupportTip: boolean = false
  headerScrolled: boolean = false
  currentTime: string = ''

  private clockTimer: number | null = null
  private mottos: string[] = [
    '愿你今天也有好心情。',
    '记录、对话、分享。',
    '有 Bug 请到社区讨论页反馈',
  ]
  private mottoIndex: number = 0

  get supportTipText(): string {
    return '已支持人数：' + this.supportCount
  }

  get headerMotto(): string {
    return this.mottos[this.mottoIndex]
  }

  created() {
    this.supportCount = parseInt(localStorage.getItem('wm_support_count') || '0', 10)
    this.supported = localStorage.getItem('wm_supported_owner') === '1'
    this.tickClock()
  }

  mounted() {
    this.clockTimer = window.setInterval(() => this.tickClock(), 1000)
    window.addEventListener('scroll', this.onWindowScroll)
  }

  beforeDestroy() {
    if (this.clockTimer) window.clearInterval(this.clockTimer)
    window.removeEventListener('scroll', this.onWindowScroll)
  }

  supportOwner() {
    if (this.supported) {
      this.showSupportTip = true
      return
    }
    this.supported = true
    this.supportCount += 1
    localStorage.setItem('wm_support_count', String(this.supportCount))
    localStorage.setItem('wm_supported_owner', '1')
    this.showSupportTip = true
  }

  nextMotto() {
    this.mottoIndex = (this.mottoIndex + 1) % this.mottos.length
  }

  tickClock() {
    const d = new Date()
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n)
    this.currentTime = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  onWindowScroll = () => {
    this.headerScrolled = window.scrollY > 8
  }
}
</script>

<style lang="stylus">
@import '~@/assets/css/index'

.site-header
  position fixed
  top 0
  left 0
  right 0
  z-index 12000
  background linear-gradient(120deg, rgba(20,30,40,.75), rgba(45,60,85,.62))
  backdrop-filter blur(7px)
  border-bottom 1px solid rgba(255,255,255,.16)
  transition all .2s ease

.site-header.scrolled
  box-shadow 0 10px 26px rgba(0,0,0,.22)
  background linear-gradient(120deg, rgba(20,30,40,.88), rgba(45,60,85,.80))

.site-header-inner
  max-width 1320px
  margin 0 auto
  display flex
  align-items center
  justify-content space-between
  gap 12px
  padding 10px 16px

.brand
  display flex
  align-items center
  gap 10px
  text-decoration none
  color #f4f8ff
  min-width 250px

.brand-avatar
  width 36px
  height 36px
  border-radius 50%
  background linear-gradient(135deg, #ffd66e, #f5a623)
  color #2b1d00
  display flex
  align-items center
  justify-content center
  font-weight 800
  box-shadow 0 6px 16px rgba(245,166,35,.38)

.brand-title
  font-size 14px
  font-weight 800

.brand-sub
  font-size 12px
  opacity .9
  max-width 320px
  white-space nowrap
  overflow hidden
  text-overflow ellipsis

.nav-group
  display flex
  align-items center
  gap 8px

.nav-btn
  border 1px solid rgba(255,255,255,.28)
  color #f4f8ff
  background rgba(255,255,255,.08)
  border-radius 999px
  padding 6px 12px
  cursor pointer
  text-decoration none
  font-size 13px
  transition transform .15s ease, background .15s ease

.nav-btn:hover
  background rgba(255,255,255,.18)
  transform translateY(-1px)

.nav-btn.router-link-exact-active
  background rgba(99,170,255,.32)
  border-color rgba(160,210,255,.5)

.nav-btn.dark
  color #1f2733
  border-color rgba(0,0,0,.2)
  background rgba(255,255,255,.92)

.support-wrap
  position relative

.nav-btn.support.done
  background rgba(80,180,120,.30)
  border-color rgba(120,220,160,.45)

.support-tip
  position absolute
  top calc(100% + 6px)
  left 0
  white-space nowrap
  background rgba(0,0,0,.85)
  color #fff
  border-radius 8px
  padding 4px 8px
  font-size 12px

.header-tools
  display flex
  align-items center
  gap 10px

.tool-btn
  border 1px dashed rgba(255,255,255,.35)
  background rgba(255,255,255,.08)
  color #f0f6ff
  border-radius 999px
  font-size 12px
  padding 6px 10px
  cursor pointer

.clock
  font-family 'Press Start 2P', monospace
  font-size 11px
  color rgba(255,255,255,.88)
  background rgba(0,0,0,.22)
  padding 6px 10px
  border-radius 8px

.modal-mask
  position fixed
  inset 0
  z-index 13000
  background rgba(0,0,0,.35)
  display flex
  align-items center
  justify-content center

.modal
  width 320px
  border-radius 12px
  background #fff
  color #222
  padding 14px

.modal-title
  font-weight 800
  margin-bottom 8px

.modal-content
  margin-bottom 12px

body
  padding-top 66px

@media (max-width: 980px)
  .site-header-inner
    flex-wrap wrap
  .brand
    min-width auto
  .brand-sub
    max-width 220px
  .header-tools
    width 100%
    justify-content flex-end

@media (max-width: 900px)
  .site-header-inner
    overflow auto
    white-space nowrap
    flex-wrap nowrap
  .brand-sub
    display none
  .header-tools
    width auto
</style>
