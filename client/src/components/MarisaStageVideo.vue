<template>
  <div class="marisa-stage-video" :class="{ mobile: mobile }" aria-hidden="true">
    <div class="marisa-stage-video__stage">
      <canvas ref="canvas" class="marisa-stage-video__canvas"></canvas>
      <video
        ref="video"
        class="marisa-stage-video__source"
        :src="src"
        muted
        playsinline
        preload="auto"
        @loadedmetadata="onLoadedMetadata"
      ></video>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'

type PlaybackMode = 'forward' | 'reverse'
type CachedFrame = {
  data: ImageData,
  time: number
}

export default Vue.extend({
  name: 'MarisaStageVideo',
  props: {
    src: {
      type: String,
      default: '/webmarisa_base.mp4'
    },
    mobile: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      mode: 'forward' as PlaybackMode,
      forwardRate: 1,
      reverseRate: 1,
      reverseLastTs: 0 as number,
      metadataReady: false,
      renderRafId: 0 as number,
      endedHandler: null as null | (() => void),
      frameCache: [] as CachedFrame[],
      lastCapturedMediaTime: -1 as number,
      cacheMaxFrames: 84 as number,
      cacheMinGap: 1 / 24 as number,
      targetWidth: 0 as number,
      targetHeight: 0 as number,
      reverseIndex: -1 as number
    }
  },
  mounted() {
    this.startForward(true)
  },
  beforeDestroy() {
    this.stopRenderLoop()
    this.detachEndedHandler()
    const video = this.getVideo()
    if (video) video.pause()
  },
  methods: {
    getVideo(): HTMLVideoElement | null {
      return this.$refs.video as HTMLVideoElement | null
    },

    getCanvas(): HTMLCanvasElement | null {
      return this.$refs.canvas as HTMLCanvasElement | null
    },

    pickRate(): number {
      return 0.96 + Math.random() * 0.08
    },

    getTargetCanvasWidth(videoWidth: number): number {
      const maxWidth = this.mobile ? 320 : 448
      return Math.max(220, Math.min(videoWidth, maxWidth))
    },

    syncCanvasSize(video: HTMLVideoElement, canvas: HTMLCanvasElement) {
      const width = this.getTargetCanvasWidth(video.videoWidth || 0)
      const height = video.videoWidth > 0
        ? Math.max(1, Math.round((video.videoHeight / video.videoWidth) * width))
        : 1

      this.targetWidth = width
      this.targetHeight = height

      if (canvas.width !== width || canvas.height !== height) {
        canvas.width = width
        canvas.height = height
      }
    },

    onLoadedMetadata() {
      const video = this.getVideo()
      const canvas = this.getCanvas()
      if (!video || !canvas) return

      this.metadataReady = true
      this.syncCanvasSize(video, canvas)
      this.frameCache = []
      this.lastCapturedMediaTime = -1
      this.reverseIndex = -1
      video.currentTime = 0
      this.renderFrame(performance.now())
      this.startForward(true)
    },

    async startForward(resetToStart: boolean = false) {
      const video = this.getVideo()
      if (!video) return

      this.mode = 'forward'
      this.reverseLastTs = 0
      this.reverseIndex = -1
      this.forwardRate = this.pickRate()
      video.playbackRate = this.forwardRate

      if (resetToStart) {
        video.currentTime = 0
        this.frameCache = []
        this.lastCapturedMediaTime = -1
      }

      this.attachEndedHandler()
      this.ensureRenderLoop()

      try {
        await video.play()
      } catch (e) {
        this.renderFrame(performance.now())
      }
    },

    startReverse() {
      const video = this.getVideo()
      if (!video || !this.metadataReady) return

      if (this.frameCache.length < 2) {
        this.startForward(true)
        return
      }

      this.mode = 'reverse'
      this.reverseRate = this.pickRate()
      this.reverseLastTs = 0
      this.reverseIndex = this.frameCache.length - 1
      this.detachEndedHandler()
      video.pause()
      this.ensureRenderLoop()
    },

    attachEndedHandler() {
      const video = this.getVideo()
      if (!video) return

      this.detachEndedHandler()
      this.endedHandler = () => {
        this.startReverse()
      }
      video.addEventListener('ended', this.endedHandler)
    },

    detachEndedHandler() {
      const video = this.getVideo()
      if (!video || !this.endedHandler) return
      video.removeEventListener('ended', this.endedHandler)
      this.endedHandler = null
    },

    ensureRenderLoop() {
      if (this.renderRafId) return
      this.renderRafId = window.requestAnimationFrame(this.renderFrame)
    },

    stopRenderLoop() {
      if (this.renderRafId) {
        window.cancelAnimationFrame(this.renderRafId)
        this.renderRafId = 0
      }
    },

    renderFrame(ts: number) {
      this.renderRafId = 0

      const video = this.getVideo()
      const canvas = this.getCanvas()
      if (!video || !canvas || !this.metadataReady) return

      const ctx = canvas.getContext('2d')
      if (!ctx) return

      this.syncCanvasSize(video, canvas)

      if (this.mode === 'reverse') {
        this.renderReverseFrame(ctx, canvas, ts)
        return
      }

      const frame = this.drawKeyedFrame(ctx, canvas, video)
      if (frame) {
        this.captureFrame(frame, video.currentTime)
      }

      this.renderRafId = window.requestAnimationFrame(this.renderFrame)
    },

    renderReverseFrame(ctx: CanvasRenderingContext2D, canvas: HTMLCanvasElement, ts: number) {
      if (!this.reverseLastTs) {
        this.reverseLastTs = ts
      }

      const gap = this.cacheMinGap / this.reverseRate
      const delta = (ts - this.reverseLastTs) / 1000

      if (delta < gap) {
        const current = this.frameCache[Math.max(0, this.reverseIndex)]
        if (current) ctx.putImageData(current.data, 0, 0)
        this.renderRafId = window.requestAnimationFrame(this.renderFrame)
        return
      }

      this.reverseLastTs = ts
      const frame = this.frameCache[Math.max(0, this.reverseIndex)]
      if (frame) {
        if (canvas.width !== frame.data.width || canvas.height !== frame.data.height) {
          canvas.width = frame.data.width
          canvas.height = frame.data.height
        }
        ctx.putImageData(frame.data, 0, 0)
      }

      this.reverseIndex -= 1
      if (this.reverseIndex <= 0) {
        this.startForward(true)
        return
      }

      this.renderRafId = window.requestAnimationFrame(this.renderFrame)
    },

    captureFrame(frame: ImageData, mediaTime: number) {
      if (mediaTime < 0) return
      if (this.lastCapturedMediaTime >= 0 && mediaTime - this.lastCapturedMediaTime < this.cacheMinGap) {
        return
      }

      this.lastCapturedMediaTime = mediaTime
      this.frameCache.push({
        data: new ImageData(new Uint8ClampedArray(frame.data), frame.width, frame.height),
        time: mediaTime
      })

      if (this.frameCache.length > this.cacheMaxFrames) {
        this.frameCache.shift()
      }
    },

    drawKeyedFrame(ctx: CanvasRenderingContext2D, canvas: HTMLCanvasElement, video: HTMLVideoElement): ImageData | null {
      if (!video.videoWidth || !video.videoHeight) return null

      ctx.clearRect(0, 0, canvas.width, canvas.height)
      ctx.drawImage(video, 0, 0, canvas.width, canvas.height)

      const frame = ctx.getImageData(0, 0, canvas.width, canvas.height)
      const data = frame.data

      for (let i = 0; i < data.length; i += 4) {
        const r = data[i]
        const g = data[i + 1]
        const b = data[i + 2]
        const minChannel = Math.min(r, g, b)
        const maxChannel = Math.max(r, g, b)
        const spread = maxChannel - minChannel
        const brightness = (r + g + b) / 3

        if (brightness >= 245 && spread <= 12) {
          data[i] = 255
          data[i + 1] = 255
          data[i + 2] = 255
          data[i + 3] = 112
        }
      }

      ctx.putImageData(frame, 0, 0)
      return frame
    }
  }
})
</script>

<style lang="stylus" scoped>
.marisa-stage-video
  position fixed
  left 50%
  bottom 118px
  transform translateX(-50%)
  z-index 9982
  pointer-events none
  width min(34vw, 416px)
  display flex
  justify-content center

.marisa-stage-video.mobile
  display none

.marisa-stage-video__stage
  position relative
  width 100%
  filter drop-shadow(0 0 10px rgba(255,255,255,.34)) drop-shadow(0 0 22px rgba(255,248,226,.26)) drop-shadow(0 18px 34px rgba(0,0,0,.22))

.marisa-stage-video__canvas
  display block
  width 100%
  height auto
  background transparent
  filter saturate(1.01) contrast(1.01)

.marisa-stage-video__source
  position absolute
  width 1px
  height 1px
  opacity 0
  pointer-events none
  left 0
  top 0

@media (max-width: 900px)
  .marisa-stage-video
    display none

  .marisa-stage-video__stage
    filter drop-shadow(0 0 8px rgba(255,255,255,.28)) drop-shadow(0 0 16px rgba(255,248,226,.22)) drop-shadow(0 14px 28px rgba(0,0,0,.18))
</style>
