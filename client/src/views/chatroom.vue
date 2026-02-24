<template>
  <div class="chatroom" :style="rootStyle">
    <div class="container">
      <!-- 左侧：自设版人格列表（仅自设版显示） -->
      <div class="persona-panel" v-if="mode === 'custom'">
        <div class="persona-header">
          <div class="persona-title">人格列表</div>
          <button class="mini-btn" @click="openCreatePersona()">+ 新建</button>
        </div>

        <div class="persona-hint" v-if="!isLoggedIn">
          自设版需要登录才能使用。
          <div style="margin-top:6px;">
            <button class="mini-btn" @click="openAuth('login')">去登录</button>
          </div>
        </div>

        <div v-else class="persona-list">
          <div
            class="persona-item"
            v-for="p in personas"
            :key="p.id"
            :class="{ active: selectedPersonaId === p.id }"
            @click="selectPersona(p)"
          >
            <div class="persona-name">{{ p.name }}</div>
            <div class="persona-actions" @click.stop>
              <button class="ui-mini" @click="openTeachModal(p)">teach表</button>
              <button class="ui-mini danger" @click="deletePersona(p)">删</button>
            </div>
          </div>

          <div class="persona-empty" v-if="personas.length === 0">
            还没有人格。点“新建”创建一个吧。
          </div>
        </div>

        <div class="persona-footer" v-if="isLoggedIn">
          <div class="persona-footer-title">自设版指令</div>
          <div class="persona-footer-line"><b>teach</b>：进入教学（下一句当“问”，再下一句当“答”）</div>
          <div class="persona-footer-line"><b>forget</b>：删除该人格最后一条 teach</div>
          <div class="persona-footer-line"><b>exit</b>：退出教学</div>
          <div class="persona-footer-line" style="opacity:.85;">规则：teach ＞ 人格 ＞ 默认设定</div>
        </div>
      </div>

      <!-- 中间：聊天区域 -->
<!-- 中间：聊天区域 -->
<div class="talk-slot">

  <!-- ✅ 拖拽判定区（拖拽时显示） -->
  <div class="dock-zones" v-if="dockEnabled && dragActive">
    <div class="zone zone-left"   :class="{ on: dragZone==='left' }"></div>
    <div class="zone zone-right"  :class="{ on: dragZone==='right' }"></div>
    <div class="zone zone-bottom" :class="{ on: dragZone==='bottom' }"></div>
    <div class="zone zone-center" :class="{ on: dragZone==='center' }"></div>
  </div>

  <div
    class="talk-panel"
    ref="talkPanel"
    :class="talkPanelClass"
    :style="talkPanelStyle"
  >
    <!-- 拖拽把手 -->
    <div
      class="dock-grip"
      v-if="dockEnabled && !dockHidden"
      @mousedown="onDockDragStart"
      @touchstart.prevent="onDockDragStart"
      title="拖动移动对话框"
    >
      ⠿
    </div>

    <!-- 隐藏按钮（非中置才显示） -->
    <button
      class="dock-hide-btn"
      v-if="dockEnabled && dockPos !== 'center' && !dockHidden"
      @click="hideDock"
      :title="'隐藏到' + dockPos"
    >
      {{ dockHideText }}
    </button>

    <!-- 顶部工具栏 -->
    <div class="topbar">
      <div class="topbar-left">
        <div class="topbar-title">うるさい! うるさい.. うるさい...</div>
        <div class="topbar-stats">
          总访问：<b>{{ stats.total_requests }}</b>
          · 总回复：<b>{{ stats.total_replies }}</b>
          · 会话：<b>{{ stats.total_sessions }}</b>
        </div>
      </div>

      <div class="topbar-right">
        <!-- 模式切换（醒目） -->
        <div class="mode-switch">
          <button
            class="mode-btn"
            :class="{ on: mode==='standard' }"
            @click="setMode('standard')"
          >标准版</button>

          <button
            class="mode-btn"
            :class="{ on: mode==='classic' }"
            @click="setMode('classic')"
            title="经典版：原汁原味的白丝魔理沙"
          >经典版</button>

          <button
            class="mode-btn"
            :class="{ on: mode==='custom' }"
            @click="setMode('custom')"
            :disabled="!isLoggedIn"
            :title="isLoggedIn ? '' : '请先登录使用自设版'"
          >自设版</button>
        </div>

        <button class="mini-btn" @click="aboutOpen = true">简介</button>

        <template v-if="isLoggedIn">
          <span class="login-badge">已登录：{{ auth.username }}</span>
          <button class="mini-btn" @click="openMessageBoard()">留言板</button>
          <button class="mini-btn danger" @click="logout()">退出</button>
        </template>

        <template v-else>
          <button class="mini-btn" @click="openAuth('login')">登录</button>
          <button class="mini-btn" @click="openAuth('register')">注册</button>
          <button class="mini-btn" @click="openMessageBoard()">留言板</button>
        </template>

        <button class="mini-btn" @click="uiOpen = !uiOpen">主题</button>
      </div>
    </div>

    <!-- ✅ 标准版：好感度显示条 -->
    <div class="affinity-bar" v-if="mode === 'standard'">
      <div class="affinity-left">
        <span class="affinity-label">好感度</span>
        <span class="affinity-value">{{ affinityText }}</span>

        <span
          class="affinity-delta"
          v-if="isLoggedInForStandard && affinityLastDelta !== 0 && affinityScore !== null"
        >
          ({{ affinityLastDelta > 0 ? '+' : '' }}{{ affinityLastDelta }})
        </span>

        <span
          class="affinity-login-hint"
          v-if="!isLoggedInForStandard"
          @click="openAuth('login')"
        >
          登录以查看
        </span>
      </div>

      <div class="affinity-right" v-if="affinityTip">
        {{ affinityTip }}
      </div>
    </div>

    <!-- 主题面板（保持你现有） -->
    <div v-if="uiOpen" class="ui-panel">
      <div class="ui-row">
        <span class="ui-label">背景</span>
        <input type="color" v-model="theme.bg" />
        <button class="ui-mini" @click="theme.bg = '#f6f2e7'">重置</button>
      </div>

      <div class="ui-row">
        <span class="ui-label">文字</span>
        <input type="color" v-model="theme.text" />
      </div>

      <div class="ui-row">
        <span class="ui-label">用户色</span>
        <input type="color" v-model="theme.you" />
      </div>

      <div class="ui-row">
        <span class="ui-label">魔理沙色</span>
        <input type="color" v-model="theme.marisa" />
      </div>

      <div class="ui-row">
        <span class="ui-label">头像样式</span>
        <select v-model="avatarMode" @change="onAvatarModePick()">
          <option value="cover">cover</option>
          <option value="contain">contain</option>
          <option value="fit-h">fit-h</option>
          <option value="fit-w">fit-w</option>
          <option value="tile">tile</option>
          <option value="pixel">pixel</option>
        </select>

        <label class="ui-check">
          <input type="checkbox" v-model="avatarLocked" />
          锁定头像
        </label>
      </div>

      <div class="ui-hint">
        说明：一旦你手动选择头像样式，会自动锁定头像（不再跟随回复切换），刷新页面后恢复默认。
      </div>
    </div>

    <div ref="talk_place" class="talk-place">
      <div
        class="talk_entry"
        v-for="(item, index) in activeTalkList"
        :class="{ 'you_color': item.name == 'You' }"
        :key="index"
      >
        <span class="talk_item" :class="{ 'you_color': item.name == 'You' }">{{ item.name }}</span>&nbsp;:&nbsp;
        <span class="talk_item" :class="{ 'you_color': item.name == 'You' }" v-html="item.content"></span>
      </div>
    </div>

    <div class="speak">
      <input
        @keydown="sendMessage($event)"
        ref="you"
        v-focus="true"
        type="text"
        name="you"
        :disabled="mode==='custom' && (!isLoggedIn || !selectedPersonaId)"
        :placeholder="inputPlaceholder"
      />
      <input
        @click="sendMessage($event)"
        ref="submit"
        type="submit"
        value="发送"
        :disabled="mode==='custom' && (!isLoggedIn || !selectedPersonaId)"
      />
    </div>
  </div> <!-- ✅ 关闭 talk-panel -->

  <!-- ✅ 展开按钮（对话框隐藏时出现在边缘） -->
  <button
    class="dock-tab"
    :class="'dock-tab--' + dockPos"
    v-if="dockEnabled && dockPos !== 'center' && dockHidden"
    :style="dockTabStyle"
    @click="showDock"
    title="展开对话框"
  >
    {{ dockTabText }}
  </button>

</div> <!-- ✅ 关闭 talk-slot -->

      <!-- 右侧：头像 + 公告 + 指令 -->
  <div class="profile" :class="{ 'dock-bottom': dockEnabled && dockPos==='bottom' }">

  <!-- ✅ 头像（底边固定时，靠左固定） -->
<div class="avatar"
     :class="{ 'avatar-fixed': dockEnabled }"
     :style="avatarStyle"></div>

  <!-- ✅ 展开按钮：放在头像上侧边沿 -->
  <button class="profile-toggle"
          :class="{ 'toggle-fixed': dockEnabled}"
          @click="toggleProfilePanel()"
          title="展开/收起公告与指令">
    ☰
  </button>

  <!-- ✅ 抽屉：公告 + 指令（默认隐藏） -->
  <div class="profile-panel"
       :class="{ 'panel-fixed': dockEnabled }"
       v-show="profilePanelOpen">

        <div class="notice">
          <div class="notice-title">📌 公告 / 社区</div>
          <div class="notice-body">
            <div>QQ群：<b>874128517</b></div>
            <div class="notice-small">（欢迎反馈 Bug / 交流用法）</div>
            <hr class="notice-hr" />
            <div class="notice-title2">使用事项</div>
            <ul class="notice-list">
              <li>请勿输入隐私信息；回答仅供娱乐/参考。</li>
              <li>默认是“自由发挥”模式，可能会有不准确内容。</li>
              <li>标准版不提供 teach；teach/forget 仅在自设版生效。</li>
              <li>留言板进入页面即预加载；未登录可看，登录后可发。</li>
            </ul>
          </div>
        </div>

        <div class="cmd">
          <span class="system-cmd">系统级指令快速说明——</span>
          <span class="system-cmd cmd-collect"><span class="marisa-cmd">status</span>&nbsp;查看目前知识所掌握情况（标准版）</span>
          <div class="cmd_desc">
            魔理沙无条件的相信你..她把你交给她的所有知识视作珍宝并会很认真的将其牢牢记住..不要让她学坏哦!
          </div>
        </div>
      </div>
    </div>

    <!-- 登录/注册弹窗（保持你原有） -->
    <div v-if="authOpen" class="modal-mask" @click.self="authOpen=false">
      <div class="modal">
        <div class="modal-header">
          <div class="tabs">
            <button class="tab" :class="{on: authTab==='login'}" @click="authTab='login'">登录</button>
            <button class="tab" :class="{on: authTab==='register'}" @click="authTab='register'">注册</button>
          </div>
          <button class="x" @click="authOpen=false">×</button>
        </div>

        <div class="modal-body">
          <div class="field">
            <div class="label">用户名</div>
            <input v-model="authForm.username" type="text" autocomplete="username" />
          </div>
          <div class="field">
            <div class="label">密码</div>
            <input v-model="authForm.password" type="password" autocomplete="current-password" />
          </div>

          <div class="hint">
            登录状态会在本机保存 <b>30 天</b>（localStorage）。
          </div>

          <div v-if="authError" class="err">{{ authError }}</div>

          <div class="actions">
            <button class="primary" @click="submitAuth()">
              {{ authTab==='login' ? '登录' : '注册' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 简介弹窗 -->
    <div v-if="aboutOpen" class="modal-mask" @click.self="aboutOpen=false">
      <div class="modal modal-wide">
        <div class="modal-header">
          <div class="modal-title">简介</div>
          <button class="x" @click="aboutOpen=false">×</button>
        </div>
        <div class="modal-body">
          <div class="about" v-html="aboutHtml"></div>
        </div>
      </div>
    </div>

    <!-- teach 表弹窗 -->
    <div v-if="teachModalOpen" class="modal-mask" @click.self="teachModalOpen=false">
      <div class="modal modal-wide">
        <div class="modal-header">
          <div class="modal-title">Teach 表：{{ teachModalTitle }}</div>
          <button class="x" @click="teachModalOpen=false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="teachLoading" class="hint">加载中…</div>
          <div v-else>
            <div v-if="teachItems.length===0" class="hint">暂无 teach</div>
            <div v-for="t in teachItems" :key="t.id" class="teach-item">
              <div class="teach-q"><b>Q：</b>{{ t.q }}</div>
              <div class="teach-a"><b>A：</b>{{ t.a }}</div>
              <div class="teach-time">{{ formatTime(t.created_at) }}</div>
              <hr class="notice-hr" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建人格弹窗 -->
    <div v-if="createPersonaOpen" class="modal-mask" @click.self="createPersonaOpen=false">
      <div class="modal modal-wide">
        <div class="modal-header">
          <div class="modal-title">新建人格</div>
          <button class="x" @click="createPersonaOpen=false">×</button>
        </div>
        <div class="modal-body">
          <div class="field">
            <div class="label">人格名称</div>
            <input v-model="createPersona.name" type="text" />
          </div>
          <div class="field">
            <div class="label">人格描述（prompt）</div>
            <textarea v-model="createPersona.prompt" class="persona-textarea"></textarea>
          </div>

          <div v-if="createPersonaError" class="err">{{ createPersonaError }}</div>

          <div class="actions">
            <button class="primary" @click="submitCreatePersona()">保存</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 留言板遮罩 -->
    <div v-if="msgOpen" class="drawer-mask" @click="closeMessageBoard()"></div>

    <!-- 留言板抽屉 -->
    <div class="drawer" :class="{open: msgOpen}">
      <div class="drawer-header">
        <div class="drawer-title">留言板</div>
        <button class="x" @click="closeMessageBoard()">×</button>
      </div>

      <div class="drawer-body">
        <div v-if="!isLoggedIn" class="hint">
          你还没登录。可以浏览留言，但只有登录用户能发表。
          <a href="javascript:void(0)" @click="openAuth('login')">去登录</a>
        </div>

        <div class="msg-list">
          <div class="msg" v-for="m in messages" :key="m.id">
            <div class="msg-meta">
              <span class="msg-user">{{ m.username }}</span>
              <span class="msg-time">{{ formatTime(m.created_at) }}</span>
            </div>
            <div class="msg-content">{{ m.content }}</div>
          </div>
        </div>
      </div>

      <div class="drawer-footer">
        <textarea v-model="newMsg" :disabled="!isLoggedIn" placeholder="说点什么…（登录后可发表）"></textarea>
        <button class="primary"
                :disabled="!isLoggedIn || newMsg.trim()==='' || msgPosting"
                @click="postMessage()">
          {{ msgPosting ? '发送中…' : '发表' }}
        </button>
      </div>
    </div>
  </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import Core from '../core';
import axios from 'axios';

const IMG_IDLE = require('./img/marisa_idle.jpg');
const IMG_THINK = require('./img/marisa_think.jpg');
const IMG_HAPPY = require('./img/marisa_happy.jpg');
const IMG_SAD = require('./img/marisa_sad.jpg');
const IMG_TEACH = require('./img/marisa_teach.jpg');
const IMG_ERROR = require('./img/marisa_error.jpg');

const MARISA: string = '白絲魔理沙';
const YOU: string = 'You';

type AvatarKey = 'idle' | 'think' | 'happy' | 'sad' | 'teach' | 'error';
type AvatarMode = 'cover' | 'contain' | 'fit-h' | 'fit-w' | 'tile' | 'pixel';

@Component({})
export default class chatroom extends Vue {
  // 标准版对话
  talk_list: any[] = [];
  // 经典版对话
  classic_talk_list: any[] = [];
  // 自设版对话
  custom_talk_list: any[] = [];

  // 三档：standard / classic / custom
  mode: 'standard' | 'classic' | 'custom' = 'standard';

  // =========================
  // ✅ 标准版好感度显示状态
  // =========================
  private affinityScore: number | null = null;
  private affinityLevel: string = '';
  private affinityCanIntimate: boolean = false;
  private affinityLastDelta: number = 0;
private affinityTimer: number | null = null;

private bgPresets: any[] = [{ key: 'plain', label: '纯色', url: '' }];
private bgKey: string = 'cnny';

private dragZone: 'left' | 'right' | 'bottom' | 'center' = 'bottom';

  // =========================
  // ✅ 凌晨锁（1:00–5:00）
  // =========================
  private lateNightLockPending: boolean = false;   // 是否正在等待用户输入“1”确认
  private lateNightPendingMsg: string = '';        // 被拦下的那一句（用于解锁后继续回答）
  private lateNightUnlockedUntil: number = 0;      // 解锁有效期（毫秒时间戳）

  private isDesktop: boolean = true;

private dockPos: 'center' | 'bottom' | 'left' | 'right' = 'bottom';
private dockHidden: boolean = false;
private dockStealth: boolean = false;

private dragActive: boolean = false;
private dragDx: number = 0;
private dragDy: number = 0;
private dragStartX: number = 0;
private dragStartY: number = 0;

private _winMove: any = null;
private _winUp: any = null;
private _winKey: any = null;
private _winResize: any = null;

private profilePanelOpen: boolean = false; // 默认隐藏

  // 自设版 teach 状态（你原来就有）
  cmd_flag: number = 0;
  teachFlag: number = 0;
  teachContent: string[] = [];

  // 经典版 teach 状态（新增）
  classicTeachMode: boolean = false;
  classicTeachBuf: string[] = [];

  avatarKey: AvatarKey = 'idle';
  avatarLocked: boolean = false;
  avatarMode: AvatarMode = 'cover';
  private avatarResetTimer: number | null = null;

  uiOpen: boolean = false;

  theme: any = {
    bg: '#f6f2e7',
    text: '#2b2b2b',
    you: '#2f6bff',
    marisa: '#1f1f1f'
  };

  stats: any = { total_requests: 0, total_replies: 0, total_sessions: 0 };
  private statsTimer: number | null = null;

  // auth
  authOpen: boolean = false;
  authTab: 'login' | 'register' = 'login';
  authForm: any = { username: '', password: '' };
  authError: string = '';
  auth: any = { token: '', username: '', exp: 0 };

  // 留言板
  msgOpen: boolean = false;
  messages: any[] = [];
  newMsg: string = '';
  msgPosting: boolean = false;

  // about
  aboutOpen: boolean = false;
  aboutHtml: string = `
    <p><b>白丝魔理沙</b>：雾雨魔理沙在外界的投影/代言人，基于 LLM + 轻量记忆 + 资料片段的互动 bot。</p>
    <p>使用提示：<b>teach</b> 教学模式（问/答两条），<b>exit</b> 退出教学；<b>status</b> 查看知识量。</p>
    <p>注意：请勿输入隐私、违法信息；回复仅供娱乐/参考。</p>
    <hr>
    <p>关于作者：</p>
    <p>我是 szcaky。《白丝魔理沙》是孙鸭子的作品，现已停服。</p>
    <p>本人在<b>友则</b>放在 GitHub 上的复制品代码：<br>
      <a href="https://github.com/TohoOutsiders/web-marisa" target="_blank" style="color: #dbb912; text-decoration: underline;">https://github.com/TohoOutsiders/web-marisa</a>
    </p>
    <p>的基础上，接通了某 AI 的 LLM，将白丝魔理沙还原。</p>
    <p>我同时也是 <b>质点安科站</b> 的站长：<br>
      <a href="https://zhidianworld.com/" target="_blank" style="color: #dbb912; text-decoration: underline;">https://zhidianworld.com/</a>
    </p>
    <p>以及安科作品 <b>詹姆斯幻想入</b> 的作者：<br>
      <a href="https://www.gululu.world/book/62226" target="_blank" style="color: #dbb912; text-decoration: underline;">https://www.gululu.world/book/62226</a>
    </p>
    <p>未来我也会尝试各种其它项目的，请多指教~</p>
  `;

  // 自设人格
  personas: any[] = [];
  selectedPersonaId: number | null = null;

  createPersonaOpen: boolean = false;
  createPersonaError: string = '';
  createPersona: any = { name: '', prompt: '' };

  teachModalOpen: boolean = false;
  teachModalTitle: string = '';
  teachLoading: boolean = false;
  teachItems: any[] = [];

  $refs!: {
    talk_place: HTMLFormElement,
    you: HTMLFormElement,
    submit: HTMLFormElement,
  }

  get activeTalkList() {
    if (this.mode === 'custom') return this.custom_talk_list;
    if (this.mode === 'classic') return this.classic_talk_list;
    return this.talk_list;
  }

  get inputPlaceholder(): string {
    if (this.mode === 'classic') {
      if (this.classicTeachMode) return '经典版教学中：先输入“问”，再输入“答”（exit 退出）';
      return '经典版：不会就输入 teach 教我（teach → 问/答两句）';
    }
    if (this.mode !== 'custom') return '输入内容...';
    if (!this.isLoggedIn) return '自设版需要先登录';
    if (!this.selectedPersonaId) return '请先在左侧选择人格';
    return '自设版输入...';
  }

  get isLoggedIn(): boolean {
    return !!this.auth.token && this.auth.exp > Date.now();
  }

  // ✅ 标准版好感度：是否已登录（允许兼容其它 token key）
  get isLoggedInForStandard(): boolean {
    const authOk = !!this.auth.token && this.auth.exp > Date.now();
    if (authOk) return true;
    const t =
      localStorage.getItem('wm_token') ||
      localStorage.getItem('token') ||
      localStorage.getItem('marisa_token') ||
      localStorage.getItem('auth_token') ||
      '';
    return !!t;
  }

  get affinityText(): string {
    if (!this.isLoggedInForStandard) return '??';
    if (this.affinityScore === null) return '...';
    return String(this.affinityScore) + '/100';
  }

  get affinityTip(): string {
    if (!this.isLoggedInForStandard) return '登录以查看';
    if (this.affinityScore === null) return '';
    if (this.affinityCanIntimate) return '好感度满啦：可解锁更亲密互动（仅轻度）';
    return '';
  }

get rootStyle() {
  var st: any = {
    backgroundColor: this.theme.bg,
    color: this.theme.text,
    '--wm-text': this.theme.text,
    '--wm-you': this.theme.you,
    '--wm-marisa': this.theme.marisa
  };

  var url = this.getBgUrlByKey(this.bgKey);
  if (url) {
    st.backgroundImage = 'url(' + url + ')';
    st.backgroundSize = 'cover';
    st.backgroundPosition = 'center';
    st.backgroundRepeat = 'no-repeat';
  } else {
    st.backgroundImage = '';
  }
  return st;
}

  get avatarStyle() {
    const map: any = {
      idle: IMG_IDLE,
      think: IMG_THINK,
      happy: IMG_HAPPY,
      sad: IMG_SAD,
      teach: IMG_TEACH,
      error: IMG_ERROR,
    };
    const img = map[this.avatarKey];

    let bgSize = 'cover';
    let bgRepeat = 'no-repeat';
    let bgPos = 'center';

    if (this.avatarMode === 'contain') bgSize = 'contain';
    if (this.avatarMode === 'fit-h') bgSize = 'auto 100%';
    if (this.avatarMode === 'fit-w') bgSize = '100% auto';
    if (this.avatarMode === 'tile') { bgSize = '64px 64px'; bgRepeat = 'repeat'; bgPos = 'left top'; }
    if (this.avatarMode === 'pixel') { bgSize = '333px'; bgRepeat = 'no-repeat'; bgPos = '58% 2%'; }

    return {
      backgroundImage: `url(${img})`,
      backgroundSize: bgSize,
      backgroundRepeat: bgRepeat,
      backgroundPosition: bgPos
    };
  }

  get dockEnabled(): boolean {
  return this.isDesktop;
}

get talkPanelClass(): any {
  return {
    'is-docked': this.dockEnabled,
    'dock-left': this.dockPos === 'left',
    'dock-right': this.dockPos === 'right',
    'dock-bottom': this.dockPos === 'bottom',
    'dock-center': this.dockPos === 'center',
    'is-hidden': this.dockHidden,
    'is-stealth': this.dockStealth
  };
}

get dockHideText(): string {
  if (this.dockPos === 'left') return '◀';
  if (this.dockPos === 'right') return '▶';
  if (this.dockPos === 'bottom') return '▼';
  return '–';
}

get dockTabText(): string {
  if (this.dockPos === 'left') return '▶';
  if (this.dockPos === 'right') return '◀';
  if (this.dockPos === 'bottom') return '▲';
  return '展开';
}

get dockTabStyle(): any {
  // 固定在对应边缘
  if (this.dockPos === 'left') {
    return { left: '6px', top: '50%', transform: 'translateY(-50%)' };
  }
  if (this.dockPos === 'right') {
    return { right: '6px', top: '50%', transform: 'translateY(-50%)' };
  }
  if (this.dockPos === 'bottom') {
    return { left: '50%', bottom: '6px', transform: 'translateX(-50%)' };
  }
  return {};
}

get talkPanelStyle(): any {
  if (!this.dockEnabled) return {};

  var st: any = { position: 'fixed', zIndex: 9990 };

  // 基础吸附位
  if (this.dockPos === 'center') {
    st.left = '50%';
    st.top = '50%';
    st.transform = 'translate(-50%,-50%)';
} else if (this.dockPos === 'bottom') {
  // ✅ bottom = 页面中间偏下（水平居中）
  st.left = '50%';
  st.bottom = '16px';
  st.transform = 'translateX(-50%)';
} else if (this.dockPos === 'left') {
    st.left = '16px';
    st.top = '50%';
    st.transform = 'translateY(-50%)';
  } else if (this.dockPos === 'right') {
    st.right = '16px';
    st.top = '50%';
    st.transform = 'translateY(-50%)';
  }

  // 拖拽偏移（拖的时候临时位移）
  if (this.dragActive) {
    st.transform = (st.transform || '') + ' translate(' + this.dragDx + 'px,' + this.dragDy + 'px)';
  }

  // 隐藏：把面板推到相邻边缘外侧，留 18px（同时禁用面板点击）
  if (this.dockHidden && this.dockPos !== 'center') {
    st.pointerEvents = 'none';
    if (this.dockPos === 'left') {
      st.transform = (st.transform || '') + ' translateX(calc(-100% + 18px))';
    } else if (this.dockPos === 'right') {
      st.transform = (st.transform || '') + ' translateX(calc(100% - 18px))';
    } else if (this.dockPos === 'bottom') {
      st.transform = (st.transform || '') + ' translateY(calc(100% - 18px))';
    }
  }

  return st;
}

  created() {
    this.loadAuthFromStorage();
  }

mounted() {
  this.fetchStats();
  this.statsTimer = window.setInterval(() => this.fetchStats(), 30000);

  // 留言板预加载
  this.fetchMessages();

  // ✅ 背景/好感度/凌晨锁
  this.syncLateNightUnlockFromStorage();
  this.loadStandardAffinity();
  this.loadBgConfig();

  // ✅ 关键：首次进入就计算一次 desktop
  this.isDesktop = window.innerWidth > 900;

  // ✅ 关键：桌面端默认下置（你想要的默认）
  if (this.isDesktop) {
    this.dockPos = 'bottom';
    this.dockHidden = false;
    this.dockStealth = false;
  } else {
    // 移动端不启用浮动 dock
    this.dockPos = 'center';
    this.dockHidden = false;
    this.dockStealth = false;
  }

  // resize：实时切换 desktop/移动端策略
  this._winResize = () => {
    this.isDesktop = window.innerWidth > 900;
    if (!this.isDesktop) {
      this.dockPos = 'center';
      this.dockHidden = false;
      this.dockStealth = false;
    } else {
      // 从移动端切回桌面端：回到底部（或你也可以保留上次位置）
      if (this.dockPos === 'center') this.dockPos = 'bottom';
    }
  };
  window.addEventListener('resize', this._winResize);

  // Shift 虚化：仅桌面端 + 不在输入框聚焦时触发
  this._winKey = (e: any) => {
    if (!this.isDesktop) return;
    if (e && e.key === 'Shift' && !e.repeat) {
      var ae: any = document.activeElement;
      var tag = ae && ae.tagName ? String(ae.tagName).toLowerCase() : '';
      if (tag === 'input' || tag === 'textarea') return;
      this.dockStealth = !this.dockStealth;
    }
  };
  window.addEventListener('keydown', this._winKey);
}

beforeDestroy() {
  if (this.statsTimer) window.clearInterval(this.statsTimer);
  if (this._winResize) window.removeEventListener('resize', this._winResize);
  if (this._winKey) window.removeEventListener('keydown', this._winKey);
}

  updated() {
    this._scrollBottom();
  }

private hideDock() {
  this.dockHidden = true;
}

private showDock() {
  this.dockHidden = false;
}

private onDockDragStart(e: any) {
  if (!this.dockEnabled) return;
  if (this.dockHidden) return;

  // 防止拖拽时选中文本
  if (e && e.preventDefault) e.preventDefault();

  var p = this.getPoint(e);
  this.dragActive = true;
  this.dragDx = 0;
  this.dragDy = 0;
  this.dragStartX = p.x;
  this.dragStartY = p.y;

  this._winMove = (ev: any) => {
  var q = this.getPoint(ev);
  this.dragDx = q.x - this.dragStartX;
  this.dragDy = q.y - this.dragStartY;

  // ✅ 实时计算命中区
  var w = window.innerWidth;
  var h = window.innerHeight;
  var x = q.x;
  var y = q.y;

  if (x < w * 0.33) this.dragZone = 'left';
  else if (x > w * 0.67) this.dragZone = 'right';
  else if (y > h * 0.72) this.dragZone = 'bottom';
  else this.dragZone = 'center';
};

this._winUp = (ev: any) => {
  var q = this.getPoint(ev);
  this.dragActive = false;

  if (this.dragZone === 'left') this.dockPos = 'left';
  else if (this.dragZone === 'right') this.dockPos = 'right';
  else if (this.dragZone === 'bottom') this.dockPos = 'bottom';
  else this.dockPos = 'center';

    this.dockHidden = false;
    this.dragDx = 0;
    this.dragDy = 0;

    window.removeEventListener('mousemove', this._winMove);
    window.removeEventListener('mouseup', this._winUp);
    window.removeEventListener('touchmove', this._winMove);
    window.removeEventListener('touchend', this._winUp);
  };

  window.addEventListener('mousemove', this._winMove);
  window.addEventListener('mouseup', this._winUp);
  window.addEventListener('touchmove', this._winMove, { passive: false } as any);
  window.addEventListener('touchend', this._winUp);
}

private getPoint(e: any): any {
  if (e && e.touches && e.touches.length) {
    return { x: e.touches[0].clientX, y: e.touches[0].clientY };
  }
  return { x: e.clientX, y: e.clientY };
}

  private setMode(m: 'standard' | 'classic' | 'custom') {
    if (m === 'custom' && !this.isLoggedIn) {
      this.authError = '请先登录以使用自设版';
      this.openAuth('login');
      return;
    }
    this.mode = m;

    // 切到标准版：尝试刷新好感度
    if (m === 'standard') {
      this.loadStandardAffinity();
    }

    // 切到自设版时加载人格列表
    if (m === 'custom') {
      this.loadPersonas();
      return;
    }

    // 离开自设版时退出自设 teach
    this.cmd_flag = 0;
    this.teachFlag = 0;
    this.teachContent = [];

    // 离开经典版时退出经典 teach
    this.classicTeachMode = false;
    this.classicTeachBuf = [];
  }

private toggleProfilePanel() {
  this.profilePanelOpen = !this.profilePanelOpen;
}

  private onAvatarModePick() {
    this.avatarLocked = true;
  }

  private setAvatar(key: AvatarKey, autoBackMs: number = 2500) {
    if (this.avatarLocked) return;
    this.avatarKey = key;

    if (this.avatarResetTimer) {
      window.clearTimeout(this.avatarResetTimer);
      this.avatarResetTimer = null;
    }

    if (autoBackMs > 0 && key !== 'teach') {
      this.avatarResetTimer = window.setTimeout(() => {
        if (this.cmd_flag !== 1) this.avatarKey = 'idle';
      }, autoBackMs);
    }
  }

  private detectAvatarFromReply(answer: string): AvatarKey {
    const s = (answer || '').toLowerCase();
    if (s.indexOf('error') >= 0 || s.indexOf('失败') >= 0 || s.indexOf('无法') >= 0) return 'error';
    if (s.indexOf('对不起') >= 0 || s.indexOf('呜呜') >= 0 || s.indexOf('抱歉') >= 0) return 'sad';
    if (s.indexOf('推理') >= 0 || s.indexOf('让我想想') >= 0 || answer.indexOf('歪头') >= 0) return 'think';
    if (answer.indexOf('✨') >= 0 || answer.indexOf('DA☆ZE') >= 0 || s.indexOf('开心') >= 0) return 'happy';
    return 'idle';
  }

  // =========================
  // ✅ 好感度：把后端字段写回状态（兼容多种字段名/结构）
  // =========================
private applyAffinityFromAnyData(data: any) {
  if (!data || typeof data !== 'object') return;

  // 兼容：data.meta / data.affinity / 直接平铺
  var raw: any = (data as any);
  if (raw.meta && typeof raw.meta === 'object') raw = raw.meta;

  var aff: any = raw;
  if (raw.affinity && typeof raw.affinity === 'object') aff = raw.affinity;

  // score
  var score: any = undefined;
  if (typeof aff.affinity_score === 'number') score = aff.affinity_score;
  else if (typeof aff.score === 'number') score = aff.score;

  if (typeof score === 'number') {
    var old: any = this.affinityScore;
    this.affinityScore = score;

    // level
    var level: any = '';
    if (typeof aff.affinity_level === 'string') level = aff.affinity_level;
    else if (typeof aff.level === 'string') level = aff.level;
    this.affinityLevel = level;

    // can_intimate
    var can: any = false;
    if (typeof aff.affinity_can_intimate === 'boolean') can = aff.affinity_can_intimate;
    else if (typeof aff.can_intimate === 'boolean') can = aff.can_intimate;
    this.affinityCanIntimate = !!can;

    if (typeof old === 'number') this.affinityLastDelta = score - old;
    else this.affinityLastDelta = 0;
  }
}

private async loadStandardAffinity() {
  if (!this.isLoggedInForStandard) {
    this.affinityScore = null;
    this.affinityLevel = '';
    this.affinityCanIntimate = false;
    this.affinityLastDelta = 0;
    return;
  }

  try {
    var fn: any = (Core as any).getMyAffinity;
    if (typeof fn !== 'function') return;

    var data: any = await fn();
    if (data) {
      this.applyAffinityFromAnyData(data);
    }
  } catch (e) { /* ignore */ }
}

  // ===== API helpers =====
  private async apiGet(path: string) {
    const headers: any = {};
    if (this.isLoggedIn) headers['Authorization'] = 'Bearer ' + this.auth.token;
    return axios.get('/api' + path, { headers: headers });
  }

  private async apiPost(path: string, data: any) {
    const headers: any = {};
    if (this.isLoggedIn) headers['Authorization'] = 'Bearer ' + this.auth.token;
    return axios.post('/api' + path, data, { headers: headers });
  }

  private async apiDelete(path: string) {
    const headers: any = {};
    if (this.isLoggedIn) headers['Authorization'] = 'Bearer ' + this.auth.token;
    return axios.delete('/api' + path, { headers: headers });
  }

  private pickErr(resp: any, fallback: string): string {
    if (!resp || !resp.data) return fallback;
    const d = resp.data;
    if (d && d.data && typeof d.data.error === 'string') return d.data.error;
    if (typeof d.error === 'string') return d.error;
    return fallback;
  }

  // ===== stats =====
  private async fetchStats() {
    try {
      const res: any = await this.apiGet('/stats');
      if (res && res.data && res.data.code === 200) {
        this.stats = res.data.data ? res.data.data : this.stats;
      }
    } catch (e) { /* ignore */ }
  }

  // ===== auth =====
  private openAuth(tab: 'login' | 'register') {
    this.authTab = tab;
    this.authOpen = true;
    // 不清空 authError，方便显示“请先登录”
    this.authForm.username = '';
    this.authForm.password = '';
  }

  private async submitAuth() {
    this.authError = '';
    const u = (this.authForm.username || '').trim();
    const p = (this.authForm.password || '').trim();
    if (u.length < 3) { this.authError = '用户名至少 3 位'; return; }
    if (p.length < 6) { this.authError = '密码至少 6 位'; return; }

    try {
      if (this.authTab === 'register') {
        const r: any = await this.apiPost('/auth/register', { username: u, password: p });
        if (!r || !r.data || r.data.code !== 200) {
          this.authError = this.pickErr(r, '注册失败');
          return;
        }
      }

      const res: any = await this.apiPost('/auth/login', { username: u, password: p });
      if (!res || !res.data || res.data.code !== 200) {
        this.authError = this.pickErr(res, '登录失败');
        return;
      }

      const token = res.data.data.token;
      const username = res.data.data.username;

      const exp = Date.now() + 30 * 24 * 60 * 60 * 1000;
      this.auth = { token: token, username: username, exp: exp };
      localStorage.setItem('wm_token', token);
      localStorage.setItem('wm_user', username);
      localStorage.setItem('wm_exp', String(exp));

      this.authOpen = false;

      // ✅ 登录成功：刷新标准版好感度
      this.loadStandardAffinity();

      // 登录成功：如果当前想用自设版，则切过去并加载 persona
      if (this.mode === 'custom') {
        this.loadPersonas();
      }
    } catch (e) {
      this.authError = '网络错误或服务未开启';
    }
  }

  private loadAuthFromStorage() {
    const token = localStorage.getItem('wm_token') || '';
    const username = localStorage.getItem('wm_user') || '';
    const exp = parseInt(localStorage.getItem('wm_exp') || '0', 10);
    if (token && username && exp > Date.now()) {
      this.auth = { token: token, username: username, exp: exp };
    } else {
      this.auth = { token: '', username: '', exp: 0 };
      localStorage.removeItem('wm_token');
      localStorage.removeItem('wm_user');
      localStorage.removeItem('wm_exp');
    }
  }

  private logout() {
    this.auth = { token: '', username: '', exp: 0 };
    localStorage.removeItem('wm_token');
    localStorage.removeItem('wm_user');
    localStorage.removeItem('wm_exp');

    // ✅ 登出后清空好感度 UI
    this.affinityScore = null;
    this.affinityLevel = '';
    this.affinityCanIntimate = false;
    this.affinityLastDelta = 0;

    // 登出后强制回标准版
    this.mode = 'standard';
    this.selectedPersonaId = null;
    this.personas = [];
  }

  // ===== 留言板 =====
  private openMessageBoard() {
    this.msgOpen = true;
    this.fetchMessages();
  }

  private closeMessageBoard() {
    this.msgOpen = false;
  }

  private async fetchMessages() {
    try {
      const res: any = await this.apiGet('/messages');
      if (res && res.data && res.data.code === 200) {
        this.messages = res.data.data ? res.data.data : [];
      }
    } catch (e) { /* ignore */ }
  }

  private async postMessage() {
    const c = this.newMsg.trim();
    if (!this.isLoggedIn) { this.openAuth('login'); return; }
    if (!c) return;

    this.msgPosting = true;
    try {
      const res: any = await this.apiPost('/messages', { content: c });
      if (res && res.data && res.data.code === 200) {
        this.newMsg = '';
        await this.fetchMessages();
      }
    } finally {
      this.msgPosting = false;
    }
  }

  // ===== 自设人格 =====
  private async loadPersonas() {
    if (!this.isLoggedIn) return;
    try {
      const res: any = await this.apiGet('/custom/personas');
      if (res && res.data && res.data.code === 200) {
        this.personas = res.data.data ? res.data.data : [];
      }
    } catch (e) { /* ignore */ }
  }

  private selectPersona(p: any) {
    this.selectedPersonaId = p.id;
    // 选择人格后：自设版输入可以用了
    this.custom_talk_list.push(Core.speak(MARISA, `（自设版）已选择人格「${p.name}」。`));
  }

  private openCreatePersona() {
    if (!this.isLoggedIn) { this.openAuth('login'); return; }
    this.createPersonaOpen = true;
    this.createPersonaError = '';
    this.createPersona.name = '';
    this.createPersona.prompt = '';
  }

  private async submitCreatePersona() {
    const name = (this.createPersona.name || '').trim();
    const prompt = (this.createPersona.prompt || '').trim();
    if (!name || !prompt) {
      this.createPersonaError = '请填写名称与描述';
      return;
    }

    try {
      const res: any = await this.apiPost('/custom/personas', { name: name, prompt: prompt });
      if (!res || !res.data || res.data.code !== 200) {
        this.createPersonaError = this.pickErr(res, '创建失败');
        return;
      }
      this.createPersonaOpen = false;
      await this.loadPersonas();
    } catch (e) {
      this.createPersonaError = '网络错误';
    }
  }

  private async deletePersona(p: any) {
    if (!confirm('确定删除人格「' + p.name + '」吗？（会连带删除 teach）')) return;
    try {
      const res: any = await this.apiDelete('/custom/personas/' + p.id);
      if (res && res.data && res.data.code === 200) {
        if (this.selectedPersonaId === p.id) this.selectedPersonaId = null;
        await this.loadPersonas();
      }
    } catch (e) { /* ignore */ }
  }

  private async openTeachModal(p: any) {
    this.teachModalTitle = p.name;
    this.teachModalOpen = true;
    this.teachLoading = true;
    this.teachItems = [];

    try {
      const res: any = await this.apiGet('/custom/personas/' + p.id + '/teach');
      if (res && res.data && res.data.code === 200) {
        this.teachItems = res.data.data ? res.data.data : [];
      }
    } catch (e) { /* ignore */ }
    this.teachLoading = false;
  }

  private tryHandleUiCommand(raw: string): boolean {
  var s = (raw || '').trim();
  if (!s) return false;

  // /bg ...
  if (s.indexOf('/bg') === 0) {
    var parts = s.split(/\s+/);
    var op = parts[1] || 'list';

    this.activeTalkList.push(Core.speak(YOU, s));

    if (op === 'list') {
      this.activeTalkList.push(Core.speak(MARISA, '可用背景：<br>' + this.bgListText()));
      return true;
    }
    if (op === 'next') {
      this.nextBg();
      this.activeTalkList.push(Core.speak(MARISA, '已切换背景：' + this.bgKey));
      return true;
    }
    // /bg xmas
    this.setBgKey(op);
    this.activeTalkList.push(Core.speak(MARISA, '已切换背景：' + this.bgKey));
    return true;
  }

  return false;
}

  // ===== chat =====
  private async sendMessage(event: KeyboardEvent | MouseEvent) {
    let _content: string = (this.$refs.you as any).value;
    _content = (_content || '').trim();
    // ✅ UI 指令（不发给后端）
if (this.tryHandleUiCommand(_content)) {
  (this.$refs.you as any).value = '';
  return;
}

    const isEnter = (event as any).keyCode === 13;
    const isClick = (event as any).button === 0;
    if (!isEnter && !isClick) return;

    if (_content === '') {
      this.activeTalkList.push(Core.speak(MARISA, 'ん？ 你说了什么咩 ¿'));
      return;
    }

    // ✅ 凌晨锁（1:00–5:00）：先提示健康，输入“1”可继续
    if (this.isLateNight()) {
      const s = _content;
      const lower = (s || '').toLowerCase();
      const emergency = this.isEmergencyMessage(lower);

      // 1) 如果正在等“1”，且用户输入了“1”：解锁，并继续回答刚才被挡的那一句
      if (this.lateNightLockPending && this.isLateNightUnlockCode(s)) {
        this.activeTalkList.push(Core.speak(YOU, s));

        this.unlockLateNight();
        this.activeTalkList.push(Core.speak(MARISA, this.lateNightUnlockedMessage()));
        if ((this as any).setAvatar) (this as any).setAvatar('happy', 1500);

        const pending = this.lateNightPendingMsg;
        this.lateNightPendingMsg = '';
        this.lateNightLockPending = false;

        // 自动继续处理上一条（不重复再显示一次 You: pending，因为上次已经显示过了）
        if (pending && pending.trim() !== '') {
          this.dispatchByMode(pending);
        }

        (this.$refs.you as any).value = '';
        return;
      }

      // 2) 未解锁且非紧急：挡住并提示（并记住这句话，等待用户输入“1”）
      if (!this.isLateNightOverrideActive() && !emergency) {
        this.activeTalkList.push(Core.speak(YOU, s));
        this.activeTalkList.push(Core.speak(MARISA, this.lateNightLockMessage()));
        if ((this as any).setAvatar) (this as any).setAvatar('sad', 2000);

        this.lateNightLockPending = true;
        this.lateNightPendingMsg = s;

        (this.$refs.you as any).value = '';
        return;
      }

      // 3) 已解锁 或 紧急消息：正常放行（继续往下走）
    } else {
      // 不在凌晨时段：清理锁状态 + 过期解锁
      this.lateNightLockPending = false;
      this.lateNightPendingMsg = '';
      this.lateNightUnlockedUntil = 0;
      localStorage.removeItem('wm_late_unlock_until');
    }

    // ✅ 自设版：未登录或未选人格 -> 直接提示（不发送到后端）
    if (this.mode === 'custom') {
      if (!this.isLoggedIn) {
        this.activeTalkList.push(Core.speak(MARISA, '自设版需要登录才能用哦～'));
        this.openAuth('login');
        return;
      }
      if (!this.selectedPersonaId) {
        this.activeTalkList.push(Core.speak(MARISA, '请先选择一个人格～'));
        return;
      }
    }

    // 先显示用户输入
    this.activeTalkList.push(Core.speak(YOU, _content));

    // 模式分流
    this.dispatchByMode(_content);

    (this.$refs.you as any).value = '';
  }

  // 标准版：保留 status，其它走默认 reply；teach/forget 在标准版不启用
  private _marisaThinkingStandard(_content: string) {
    if (_content === 'status') { this._marisaStatus(); return; }

    // 标准版：不使用 teach/forget
    if (_content === 'teach' || _content === 'forget') {
      this.activeTalkList.push(Core.speak(MARISA, '标准版不支持 teach/forget，请切换到自设版使用～'));
      return;
    }

    this._marisaReplyStandard(_content);
  }

  // ✅ 标准版：走 Core.replyStandard（带 token + 返回 meta/affinity）
  private async _marisaReplyStandard(_content: string) {
    this.setAvatar('think', 0);

    const data = await (Core as any).replyStandard(_content);

    // 兼容：data.answer
    if (data && typeof data.answer === 'string') {
      const answer = data.answer;

      this.activeTalkList.push(Core.speak(MARISA, answer));
      this.setAvatar(this.detectAvatarFromReply(answer), 2500);

      // ✅ 更新好感度显示
      this.applyAffinityFromAnyData(data);

      // ✅ 达到 100 给提示（只提示；真正放行逻辑仍应后端控制）
      if (this.affinityCanIntimate) {
        this.activeTalkList.push(Core.speak(MARISA, '（小声）…好感度满了诶。只允许轻轻的那种，别得寸进尺DA☆ZE！'));
      }
      return;
    }

    // 兼容：如果后端还是只回字符串（极少数情况）
    if (typeof data === 'string') {
      this.activeTalkList.push(Core.speak(MARISA, data));
      this.setAvatar(this.detectAvatarFromReply(data), 2500);
      return;
    }

    // fallback
    this.activeTalkList.push(Core.speak(MARISA, '（标准版）网络错误…'));
    this.setAvatar('error', 3500);
  }

  // ===== 经典版：teach/forget/status + /api/classic/reply =====
  private _marisaThinkingClassic(_content: string) {
    // teach 进入教学
    if (_content === 'teach') {
      this.classicTeachMode = true;
      this.classicTeachBuf = [];
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）教学模式启动！下一句当“问”，再下一句当“答”。输入 exit 退出。'));
      this.setAvatar('teach', 0);
      return;
    }

    // forget 删除最后一条
    if (_content === 'forget') {
      this._classicForget();
      return;
    }

    // status 查看经典库条目数
    if (_content === 'status') {
      this._classicStatus();
      return;
    }

    // 教学流程
    if (this.classicTeachMode) {
      this._classicTeachFlow(_content);
      return;
    }

    // 普通对话
    this._classicReply(_content);
  }

  private async _classicReply(q: string) {
    this.setAvatar('think', 0);
    try {
      const res: any = await this.apiPost('/classic/reply', { keyword: q });
      if (res && res.data && res.data.code === 200 && res.data.data && res.data.data.answer) {
        const ans = res.data.data.answer;
        this.classic_talk_list.push(Core.speak(MARISA, ans));
        this.setAvatar(this.detectAvatarFromReply(ans), 2500);
      } else {
        this.classic_talk_list.push(Core.speak(MARISA, '（经典版）我没听懂…你教教我吧？teach！'));
        this.setAvatar('think', 1500);
      }
    } catch (e) {
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）网络错误…'));
      this.setAvatar('error', 2500);
    }
  }

  private async _classicStatus() {
    try {
      const res: any = await this.apiGet('/classic/status');
      if (res && res.data && res.data.code === 200 && res.data.data) {
        const n = res.data.data.count;
        this.classic_talk_list.push(Core.speak(MARISA, `（经典版）我现在记住了 ${n} 条问答…继续教我嘛！`));
      }
    } catch (e) {
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）查不到脑重量…'));
    }
  }

  private async _classicForget() {
    try {
      const res: any = await this.apiPost('/classic/forget', {});
      if (res && res.data && res.data.code === 200) {
        const ok = res.data.data && res.data.data.ok;
        this.classic_talk_list.push(Core.speak(MARISA, ok ? '（经典版）已忘记最后一条！' : '（经典版）没东西可忘了…'));
      }
    } catch (e) {
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）忘记失败…'));
    }
  }

  private async _classicTeachFlow(_content: string) {
    if (_content === 'exit') {
      this.classicTeachMode = false;
      this.classicTeachBuf = [];
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）退出教学模式。'));
      this.setAvatar('idle', 0);
      return;
    }

    this.classicTeachBuf.push(_content);

    if (this.classicTeachBuf.length === 1) {
      this.classic_talk_list.push(Core.speak(MARISA, '（经典版）好，那“答”是什么？'));
      return;
    }

    if (this.classicTeachBuf.length >= 2) {
      const q = this.classicTeachBuf[0];
      const a = this.classicTeachBuf[1];

      try {
        const res: any = await this.apiPost('/classic/teach', { answer: q + '`' + a });
        if (res && res.data && res.data.code === 200) {
          const id = res.data.data && res.data.data.id;
          this.classic_talk_list.push(Core.speak(MARISA, `（经典版）行，我记住啦！（id=${id}）`));
        } else {
          this.classic_talk_list.push(Core.speak(MARISA, '（经典版）我没记住…再来一次？'));
        }
      } catch (e) {
        this.classic_talk_list.push(Core.speak(MARISA, '（经典版）网络错误，没记住…'));
      }

      this.classicTeachMode = false;
      this.classicTeachBuf = [];
      this.setAvatar('idle', 0);
    }
  }

  private async loadBgConfig() {
  // 读本地保存
  var saved = localStorage.getItem('wm_bg_key');
  if (saved) this.bgKey = saved;

  // 拉 manifest（失败就用默认）
  try {
    var resp = await fetch('/bg/manifest.json', { cache: 'no-store' } as any);
    if (!resp.ok) return;
    var arr = await resp.json();
    if (arr && arr.length) {
      this.bgPresets = arr;
      // 如果保存的 key 不存在，回退到第一个
      if (!this.getBgUrlByKey(this.bgKey) && this.bgKey !== 'plain') {
        this.bgKey = arr[0].key || 'plain';
      }
    }
  } catch (e) { /* ignore */ }
}

private getBgUrlByKey(key: string): string {
  for (var i = 0; i < this.bgPresets.length; i++) {
    if (this.bgPresets[i].key === key) return this.bgPresets[i].url || '';
  }
  return '';
}

private setBgKey(key: string) {
  this.bgKey = key;
  localStorage.setItem('wm_bg_key', key);
}

private bgListText(): string {
  var out: string[] = [];
  for (var i = 0; i < this.bgPresets.length; i++) {
    var it = this.bgPresets[i];
    out.push(it.key + '：' + (it.label || it.key));
  }
  return out.join('<br>');
}

private nextBg() {
  if (!this.bgPresets || this.bgPresets.length === 0) return;
  var idx = 0;
  for (var i = 0; i < this.bgPresets.length; i++) {
    if (this.bgPresets[i].key === this.bgKey) { idx = i; break; }
  }
  var next = this.bgPresets[(idx + 1) % this.bgPresets.length];
  this.setBgKey(next.key);
}

  // 自设版：teach/forget 生效；普通对话走 /api/custom/reply
  private _marisaThinkingCustom(_content: string) {
    if (_content === 'teach') {
      this.custom_talk_list.push(Core.speak(MARISA, '（自设版）教学模式启动！下一句当“问”，再下一句当“答”。退出输入 exit。'));
      this.cmd_flag = 1;
      this.teachFlag = 0;
      this.teachContent = [];
      this.setAvatar('teach', 0);
      return;
    }

    if (_content === 'forget') {
      this._customForgetLastTeach();
      return;
    }

    if (this.cmd_flag === 1) {
      this._customTeachFlow(_content);
      return;
    }

    this._customReply(_content);
  }

  private async _customReply(_content: string) {
    this.setAvatar('think', 0);
    try {
      const res: any = await this.apiPost('/custom/reply', { persona_id: this.selectedPersonaId, message: _content });
      if (res && res.data && res.data.code === 200) {
        const ans = res.data.data.answer;
        this.custom_talk_list.push(Core.speak(MARISA, ans));
        this.setAvatar(this.detectAvatarFromReply(ans), 2500);
      } else {
        this.custom_talk_list.push(Core.speak(MARISA, '自设版回复失败了…'));
        this.setAvatar('error', 3500);
      }
    } catch (e) {
      this.custom_talk_list.push(Core.speak(MARISA, '自设版网络错误…'));
      this.setAvatar('error', 3500);
    }
  }

  // teach：两句（Q/A）后写入 persona teach 表
  private async _customTeachFlow(_content: string) {
    if (_content === 'exit') {
      this.custom_talk_list.push(Core.speak(MARISA, '（自设版）退出教学模式。'));
      this.cmd_flag = 0;
      this.teachFlag = 0;
      this.teachContent = [];
      this.setAvatar('idle', 0);
      return;
    }

    this.teachFlag++;
    this.teachContent.push(_content);

    if (this.teachFlag === 1) {
      this.custom_talk_list.push(Core.speak(MARISA, '好，那么“答”是什么？'));
      return;
    }

    if (this.teachFlag >= 2) {
      const q = this.teachContent[0];
      const a = this.teachContent[1];

      try {
        const url = '/custom/personas/' + this.selectedPersonaId + '/teach';
        const res: any = await this.apiPost(url, { q: q, a: a });
        if (res && res.data && res.data.code === 200) {
          this.custom_talk_list.push(Core.speak(MARISA, '行，我知道了（已写入该人格 teach）'));
        } else {
          this.custom_talk_list.push(Core.speak(MARISA, '写入 teach 失败…'));
        }
      } catch (e) {
        this.custom_talk_list.push(Core.speak(MARISA, '写入 teach 网络错误…'));
      }

      this.cmd_flag = 0;
      this.teachFlag = 0;
      this.teachContent = [];
      this.setAvatar('idle', 0);
    }
  }

  // forget：删最后一条 teach
  private async _customForgetLastTeach() {
    if (!this.selectedPersonaId) return;
    try {
      const res: any = await this.apiDelete('/custom/personas/' + this.selectedPersonaId + '/teach/last');
      if (res && res.data && res.data.code === 200) {
        this.custom_talk_list.push(Core.speak(MARISA, '（自设版）已删除最后一条 teach'));
      } else {
        this.custom_talk_list.push(Core.speak(MARISA, '（自设版）删除失败…'));
      }
    } catch (e) {
      this.custom_talk_list.push(Core.speak(MARISA, '（自设版）网络错误…'));
    }
  }

  // 标准版 status：仍用旧接口
  private async _marisaStatus() {
    const weight: number = await Core.status();
    if (weight) this.activeTalkList.push(Core.speak(MARISA, `目前魔理沙的脑重量是 ${weight} 克。`));
    else this.activeTalkList.push(Core.speak(MARISA, `我的记忆要一片混乱了 ...`));
  }

  private formatTime(ts: number) {
    if (!ts) return '';
    const d = new Date(ts * 1000);
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n);
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  private _scrollBottom() {
    this.$nextTick(() => {
      const tp: any = this.$refs.talk_place as any;
      if (tp) tp.scrollTop = tp.scrollHeight;
    });
  }

  private isLateNight(): boolean {
    const h = new Date().getHours(); // 用户设备本地时间
    return h >= 1 && h < 5;
  }

  private lateNightReply(): string {
    return "（压低声音）喂喂…现在都凌晨了诶！先去睡觉啦，醒来再来找我～( ˘ω˘ )";
  }

  private syncLateNightUnlockFromStorage() {
    const v = parseInt(localStorage.getItem('wm_late_unlock_until') || '0', 10);
    if (v > Date.now()) {
      this.lateNightUnlockedUntil = v;
    } else {
      this.lateNightUnlockedUntil = 0;
      localStorage.removeItem('wm_late_unlock_until');
    }
  }

  private isLateNightUnlockCode(s: string): boolean {
    // 兼容全角 １
    return s === '1' || s === '１';
  }

  private computeLateNightUnlockUntil(): number {
    const now = new Date();
    const until = new Date(now);
    // 解锁到本地时间 05:00
    until.setHours(5, 0, 0, 0);

    // 理论上只会在 1~5 点触发，这里做个兜底
    if (until.getTime() <= now.getTime()) {
      until.setDate(until.getDate() + 1);
    }
    return until.getTime();
  }

  private isLateNightOverrideActive(): boolean {
    return this.lateNightUnlockedUntil > Date.now();
  }

  private unlockLateNight() {
    const until = this.computeLateNightUnlockUntil();
    this.lateNightUnlockedUntil = until;
    localStorage.setItem('wm_late_unlock_until', String(until));

    this.lateNightLockPending = false;
    // 不清 pendingMsg：下面会取出来继续处理，再清
  }

  private lateNightLockMessage(): string {
    return '现在已经是凌晨了哦，为了您的健康着想，还是先去睡觉，明天再来找白丝魔理沙吧！'
      + '<br>（如果您执意继续使用，请输入“1”，白丝魔理沙将恢复可用）';
  }

  private lateNightUnlockedMessage(): string {
    // 显示解锁到几点（通常 05:00）
    const d = new Date(this.lateNightUnlockedUntil);
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n);
    const hh = pad(d.getHours());
    const mm = pad(d.getMinutes());
    return '……好吧。既然你坚持，那我就继续陪你一会儿。'
      + `<br><span style="opacity:.85">（已临时解锁，持续到 ${hh}:${mm}）</span>`;
  }

  private dispatchByMode(_content: string) {
    if (this.mode === 'standard') {
      this._marisaThinkingStandard(_content);
    } else if (this.mode === 'classic') {
      this._marisaThinkingClassic(_content);
    } else {
      this._marisaThinkingCustom(_content);
    }
  }

  private isEmergencyMessage(sLower: string): boolean {
    // 你原来的紧急词表：保留，仍然允许直接通过不需要解锁
    return (
      sLower.indexOf('紧急') >= 0 ||
      sLower.indexOf('救命') >= 0 ||
      sLower.indexOf('报警') >= 0 ||
      sLower.indexOf('自杀') >= 0 ||
      sLower.indexOf('suicide') >= 0 ||
      sLower.indexOf('kill myself') >= 0
    );
  }

}
</script>

<style lang="stylus" scoped>
@import './index'

/* 让默认文字色/消息色可控 */
.talk_entry
  color var(--wm-marisa)

.you_color
  color var(--wm-you) !important

/* ✅ 标准版：好感度条 */
.affinity-bar
  display flex
  align-items center
  justify-content space-between
  padding 8px 10px
  margin 8px 10px
  border-radius 12px
  background rgba(0,0,0,0.06)
  border 1px solid rgba(0,0,0,.10)
  font-size 12px

.affinity-left
  display flex
  align-items center
  gap 8px

.affinity-label
  opacity 0.8

.affinity-value
  font-weight 800

.affinity-delta
  opacity .75

.affinity-login-hint
  color #2b6fff
  cursor pointer

.affinity-right
  opacity 0.85

/* 左侧人格面板 */
.persona-panel
  width 220px
  height 502px
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  background rgba(255,255,255,.78)
  padding 10px
  display flex
  flex-direction column

.persona-header
  display flex
  justify-content space-between
  align-items center
  margin-bottom 8px

.persona-title
  font-weight 800
  color rgba(0,0,0,.8)

.persona-hint
  font-size 12px
  color rgba(0,0,0,.65)
  line-height 1.5

.persona-list
  flex 1
  overflow auto
  padding-right 4px

.persona-item
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  padding 8px
  margin-bottom 8px
  background rgba(255,255,255,.7)
  cursor pointer

.persona-item.active
  border-color rgba(60,120,255,.35)
  background rgba(210,235,255,.85)

.persona-name
  font-weight 800
  color rgba(0,0,0,.8)
  margin-bottom 6px

.persona-actions
  display flex
  gap 6px

.persona-empty
  font-size 12px
  opacity .7
  padding 10px

.persona-footer
  margin-top 8px
  font-size 12px
  color rgba(0,0,0,.7)

.persona-footer-title
  font-weight 800
  margin-bottom 6px

.persona-footer-line
  margin 4px 0

/* 模式切换 */
.mode-switch
  display flex
  border 1px solid rgba(60,120,255,.25)
  border-radius 10px
  overflow hidden

.mode-btn
  padding 5px 10px
  background rgba(220,235,255,.65)
  border none
  cursor pointer
  font-size 12px
  font-weight 800
  color rgba(30,90,210,1)

.mode-btn.on
  background rgba(60,120,255,.85)
  color #fff

.mode-btn:disabled
  opacity .4
  cursor not-allowed

/* 顶部工具栏 */
.topbar
  display flex
  justify-content space-between
  align-items center
  padding 8px 10px
  background rgba(20, 30, 40, 0.12)
  border-bottom 1px solid rgba(0,0,0,.12)

.topbar-left
  display flex
  flex-direction column
  gap 4px

.topbar-title
  font-size 12px
  color rgba(0,0,0,.7)

.topbar-stats
  font-size 12px
  color rgba(0,0,0,.78)

.topbar-right
  display flex
  align-items center
  gap 6px

.mini-btn
  padding 5px 10px
  border 1px solid rgba(60,120,255,.35)
  background rgba(220,235,255,.85)
  color rgba(30,90,210,1)
  border-radius 10px
  cursor pointer
  font-size 12px
  font-weight 700

.mini-btn:hover
  background rgba(220,235,255,1)

.mini-btn.danger
  border-color rgba(255,80,80,.35)
  background rgba(255,230,230,.85)
  color rgba(200,40,40,1)

.login-badge
  font-size 12px
  color rgba(0,0,0,.75)
  background rgba(255,255,255,.75)
  border 1px solid rgba(0,0,0,.08)
  padding 4px 8px
  border-radius 10px

/* 主题面板 */
.ui-panel
  margin 8px 10px
  padding 10px
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  background rgba(255,255,255,.78)

.ui-row
  display flex
  align-items center
  gap 10px
  margin 6px 0

.ui-label
  width 72px
  font-size 12px
  color rgba(0,0,0,.75)

.ui-mini
  padding 4px 8px
  border 1px solid rgba(0,0,0,.12)
  border-radius 10px
  background rgba(255,255,255,.85)
  cursor pointer
  font-size 12px

.ui-mini.danger
  border-color rgba(255,80,80,.35)
  color rgba(200,40,40,1)

.ui-check
  display flex
  align-items center
  gap 6px
  font-size 12px
  color rgba(0,0,0,.75)

.ui-hint
  margin-top 8px
  font-size 12px
  color rgba(0,0,0,.65)

.persona-textarea
  width 100%
  min-height 140px
  resize vertical
  padding 8px
  border-radius 10px
  border 1px solid rgba(0,0,0,.12)
  background rgba(255,255,255,.85)
  outline none

.teach-item
  margin 8px 0

.teach-q, .teach-a
  margin 4px 0

.teach-time
  font-size 12px
  opacity .7

/* 弹窗 */
.modal-mask
  position fixed
  top 0
  left 0
  right 0
  bottom 0
  background rgba(0,0,0,.35)
  display flex
  align-items center
  justify-content center
  z-index 9999

.modal
  width 420px
  background rgba(240,250,250,.98)
  border 1px solid rgba(0,0,0,.12)
  border-radius 14px
  color rgba(0,0,0,.85)
  overflow hidden

.modal-wide
  width 720px
  max-width 92vw

.modal-header
  display flex
  justify-content space-between
  align-items center
  padding 10px 12px
  border-bottom 1px solid rgba(0,0,0,.12)

.modal-title
  font-weight 800

.x
  background transparent
  border none
  color rgba(0,0,0,.7)
  font-size 18px
  cursor pointer

.modal-body
  padding 12px

.tabs
  display flex
  gap 8px

.tab
  padding 6px 10px
  border 1px solid rgba(0,0,0,.12)
  border-radius 10px
  background rgba(255,255,255,.75)
  color rgba(0,0,0,.75)
  cursor pointer
  font-size 12px

.tab.on
  background rgba(210,235,255,.9)
  border-color rgba(60,120,255,.35)
  color rgba(30,90,210,1)

.field
  margin 10px 0

.label
  font-size 12px
  opacity .85
  margin-bottom 4px

.field input
  width 100%
  padding 8px 10px
  border-radius 10px
  border 1px solid rgba(0,0,0,.12)
  background rgba(255,255,255,.85)
  color rgba(0,0,0,.85)
  outline none

.hint
  font-size 12px
  opacity .85
  margin-top 8px

.err
  margin-top 10px
  color #c62828
  font-size 12px

.actions
  margin-top 12px
  display flex
  justify-content flex-end

.primary
  padding 8px 12px
  border-radius 10px
  border 1px solid rgba(60,120,255,.35)
  background rgba(210,235,255,.9)
  color rgba(30,90,210,1)
  cursor pointer
  font-weight 700

.primary:disabled
  opacity .5
  cursor not-allowed

.about
  font-size 13px
  line-height 1.7
  color rgba(0,0,0,.82)

/* 留言板遮罩 */
.drawer-mask
  position fixed
  top 0
  left 0
  right 0
  bottom 0
  background rgba(0,0,0,.18)
  z-index 9997

/* 留言板抽屉 */
.drawer
  position fixed
  top 10px
  right -460px
  width 440px
  height calc(100vh - 20px)
  background rgba(220,245,245,.60)
  backdrop-filter blur(6px)
  border 1px solid rgba(0,0,0,.12)
  border-radius 14px
  z-index 9998
  transition right .2s ease, background .2s ease
  display flex
  flex-direction column
  color rgba(0,0,0,.85)

.drawer.open
  right 12px
  background rgba(220,245,245,.95)

.drawer-header
  padding 10px 12px
  border-bottom 1px solid rgba(0,0,0,.12)
  display flex
  justify-content space-between
  align-items center

.drawer-title
  font-weight 800

.drawer-body
  padding 10px 12px
  overflow auto
  flex 1

.msg
  padding 10px
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  margin-bottom 10px
  background rgba(255,255,255,.75)

.msg-meta
  display flex
  justify-content space-between
  font-size 12px
  opacity .85
  margin-bottom 6px

.msg-content
  font-size 13px
  line-height 1.6
  white-space pre-wrap

.drawer-footer
  padding 10px 12px
  border-top 1px solid rgba(0,0,0,.12)
  display flex
  gap 8px

.drawer-footer textarea
  flex 1
  height 64px
  resize none
  padding 8px
  border-radius 10px
  border 1px solid rgba(0,0,0,.15)
  background rgba(255,255,255,.85)
  color rgba(0,0,0,.9)
  outline none

.drawer-footer textarea:disabled
  background rgba(255,255,255,.55)
  color rgba(0,0,0,.55)

/* 公告 */
.notice
  margin-top 12px
  padding 12px
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  background rgba(255,255,255,.65)
  color rgba(0,0,0,.82)
  font-size 13px
  line-height 1.6

.notice-title
  font-weight 800
  margin-bottom 6px

.notice-title2
  font-weight 800
  margin 8px 0 4px 0

.notice-small
  font-size 12px
  opacity .85

.notice-hr
  border none
  border-top 1px solid rgba(0,0,0,.12)
  margin 10px 0

.notice-list
  padding-left 18px
  margin 0

.notice-list li
  margin 4px 0

/* ========== Mobile Responsive ========== */
@media (max-width: 900px)
  .chatroom
    position: relative !important
    top: auto !important
    left: auto !important
    transform: none !important
    width: 100% !important
    height: auto !important
    min-height: 100vh
    padding: 10px
    box-sizing: border-box

  .container
    width: 100% !important
    height: auto !important
    flex-direction: column
    gap: 10px

  .persona-panel
    width: 100% !important
    height: auto !important
    max-height: 240px
    overflow-y: auto
    -webkit-overflow-scrolling: touch

  .talk-panel
    width: 100% !important
    height: auto !important
    min-height: 60vh

  .topbar
    flex-direction: column
    align-items: stretch
    gap: 8px

  .topbar-right
    flex-wrap: wrap
    gap: 8px

  .mode-switch
    width: 100%
    justify-content: flex-start

  .affinity-bar
    margin 8px 0

  .talk-place
    max-height: 55vh
    overflow-y: auto
    -webkit-overflow-scrolling: touch

  .speak
    height: auto
    padding: 8px
    gap: 8px

  .speak input[name='you']
    height: 38px
    font-size: 16px
    border-radius: 10px

  .speak input[type='submit']
    width: 84px
    height: 38px
    font-size: 14px

  .profile
    width: 100% !important
    height: auto !important

  .avatar
    width: 100% !important
    height: 220px !important
    background-position: center !important
    margin-bottom: 10px

  .notice, .cmd
    width: 100% !important

  .drawer
    width: 100% !important
    right: 0 !important
    left: 0 !important
    top: auto !important
    bottom: -110vh
    height: 75vh
    border-radius: 16px 16px 0 0
    transition: bottom .2s ease, background .2s ease

  .drawer.open
    bottom: 0

  .drawer-body
    -webkit-overflow-scrolling: touch
    overflow-y: auto

.talk-slot
  position: relative
  width: 700px
  height: 500px

/* 浮层时更像“对话框” */
.talk-panel.is-docked
  border-radius 14px
  box-shadow 0 16px 40px rgba(0,0,0,.25)

/* Shift 虚化隐藏 */
.talk-panel.is-stealth
  opacity .08
  filter blur(6px)
  pointer-events none
  transition opacity .18s ease, filter .18s ease

.dock-grip
  position absolute
  top 8px
  left 8px
  width 26px
  height 26px
  border-radius 8px
  display flex
  align-items center
  justify-content center
  background rgba(255,255,255,.65)
  border 1px solid rgba(0,0,0,.12)
  cursor grab
  user-select none
  z-index 10

.dock-hide-btn
  position absolute
  z-index 10
  border none
  background rgba(255,255,255,.75)
  border 1px solid rgba(0,0,0,.12)
  cursor pointer
  color rgba(0,0,0,.7)
  border-radius 10px

/* 左/右/下：隐藏按钮放在“朝向页面中心”的那侧 */
.talk-panel.dock-left .dock-hide-btn
  right -10px
  top 50%
  width 18px
  height 46px
  transform translateY(-50%)

.talk-panel.dock-right .dock-hide-btn
  left -10px
  top 50%
  width 18px
  height 46px
  transform translateY(-50%)

.talk-panel.dock-bottom .dock-hide-btn
  top -10px
  left 50%
  width 46px
  height 18px
  transform translateX(-50%)

.dock-tab
  position fixed
  z-index 9991
  border none
  background rgba(255,255,255,.82)
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  cursor pointer
  width 26px
  height 64px
  color rgba(0,0,0,.7)

  /* ✅ 拖放判定区 overlay */
.dock-zones
  position fixed
  inset 0
  z-index 9988
  pointer-events none

.dock-zones .zone
  position absolute
  border 2px dashed rgba(0,0,0,.35)
  background rgba(255,255,255,.06)

.dock-zones .zone.on
  border-color rgba(60,120,255,.95)
  background rgba(60,120,255,.10)

.dock-zones .zone-left
  left 10px
  top 10px
  bottom 10px
  width 33vw
  border-radius 14px

.dock-zones .zone-right
  right 10px
  top 10px
  bottom 10px
  width 33vw
  border-radius 14px

.dock-zones .zone-bottom
  left 10px
  right 10px
  bottom 10px
  height 28vh
  border-radius 14px

.dock-zones .zone-center
  left 34vw
  right 34vw
  top 10px
  bottom 30vh
  border-radius 14px

/* ✅ dock-tab 形态：左右竖条 / 底部横条 */
.dock-tab
  position fixed
  z-index 9991
  border none
  background rgba(255,255,255,.85)
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  cursor pointer
  color rgba(0,0,0,.7)

.dock-tab--left,
.dock-tab--right
  width 26px
  height 64px

.dock-tab--bottom
  width 64px
  height 26px

/* ✅ 底边固定头像 + 公告抽屉 */
/* ✅ 头像固定在：对话框(bottom居中)的左侧，并贴住底边 */
.avatar.avatar-fixed
  position fixed
  left calc(50% - 608px)
  bottom 16px
  z-index 9992

.profile-toggle
  border none
  background rgba(255,255,255,.85)
  border 1px solid rgba(0,0,0,.12)
  border-radius 12px
  cursor pointer
  color rgba(0,0,0,.75)
  width 28px
  height 28px

.profile-toggle.toggle-fixed
  position fixed
  left calc(50% - 608px)
  bottom calc(16px + 250px - 14px)
  z-index 9993

.profile-panel.panel-fixed
  position fixed
  left calc(50% - 608px)
  bottom calc(16px + 250px + 10px)
  width 250px
  z-index 9992

@media (max-width: 1250px)
  .avatar.avatar-fixed
    left 16px
  .profile-toggle.toggle-fixed
    left 16px
  .profile-panel.panel-fixed
    left 16px

</style>
