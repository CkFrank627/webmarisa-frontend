// client/src/core/index.ts
import Api from '../api/v1'
import Tools from '../assets/js/tools'

interface ISpeakConfig {
  name: string
  content: string
}

export default class Core {
  /**
   * 魔理沙说话格式,以及处理You的说话格式
   */
  public static speak (name: string, content: string) : Object {
    const obj: ISpeakConfig = { name, content }
    return obj
  }

  // ===== 新增：token & headers =====
  private static getToken(): string {
    // 你现在登录保存的是 wm_token（chatroom.vue 里 setItem('wm_token', token)）
    return (
      localStorage.getItem('wm_token') ||
      localStorage.getItem('token') ||
      localStorage.getItem('marisa_token') ||
      localStorage.getItem('auth_token') ||
      ''
    )
  }

  private static authHeaders(): any {
    const t = Core.getToken()
    if (!t) return {}
    return { Authorization: 'Bearer ' + t }
  }

  /**
   * ✅ 新增：标准版 reply（返回 data：包含 answer + affinity_*）
   * 后端返回结构：{ code: 200, data: { answer: "...", affinity_score, affinity_level, affinity_can_intimate } }
   */
  public static async replyStandard (content: string): Promise<any | undefined> {
    try {
      const resp = await fetch('/Reply', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...Core.authHeaders(),
        } as any,
        body: JSON.stringify({ keyword: content }),
        credentials: 'include', // ✅ 保留 cookie（sid 会话记忆）
      })

      const json = await resp.json()
      if (json && json.data) return json.data
      return undefined
    } catch (err) {
      // 注意：这里不要 throw，保持兼容旧逻辑
      // eslint-disable-next-line no-console
      console.log(`回复失败 ... ${err}`)
      return undefined
    }
  }

  /**
   * 回复逻辑判断中枢（旧接口兼容：只返回 answer string）
   */
  public static async reply (content: string) : Promise<any> {
    const data = await Core.replyStandard(content)
    if (data && typeof data.answer === 'string') return data.answer

    // 兼容：如果你旧 Api.fecthMemory 仍能用，也可 fallback
    try {
      const config = { keyword: content }
      const res = await Api.fecthMemory(config as any)
      return res.data.data.answer
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(`回复失败 ... ${err}`)
      return undefined
    }
  }

  /**
   * （可选）获取当前用户好感度
   * 需要你后端存在 GET /api/affinity/me
   */
  public static async getMyAffinity(): Promise<any | undefined> {
    try {
      const resp = await fetch('/api/affinity/me', {
        method: 'GET',
        headers: {
          ...Core.authHeaders(),
        } as any,
        credentials: 'include',
      })
      const json = await resp.json()
      if (json && json.data) return json.data
      return undefined
    } catch (err) {
      return undefined
    }
  }

  /**
   * 学习中枢
   */
  public static async teach (content: string) : Promise<any> {
    const str = content.split('')
    const realIp: string = await Tools.getIp()
    const config = {
      ip: realIp,
      keyword: str[0],
      answer: str[1]
    }

    try {
      const res = await Api.AddMemory(config as any)
      if (res.data.data.code === 200) {
        return true
      }
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(`无法学习 ... ${err}`)
      return false
    }
  }

  /**
   * 记忆消除中枢
   */
  public static async forget (list: any[]) : Promise<any> {
    const len: number = list.length
    let answer: string = list[1].content
    if (len > 3) answer = list[len - 2].content

    const config = { answer }

    try {
      const res = await Api.DeleteMemoryByAnswer(config as any)
      if (res.data.code === 200 && res.data.data === 'success') {
        return true
      }
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(`无法忘记 ... ${err}`)
      return false
    }
  }

  /**
   * 记忆重量
   */
  public static async status() : Promise<any> {
    try {
      const res = await Api.FecthMemoryCount()
      if (res.data.code === 200 && res.data.hasOwnProperty('data')) {
        return res.data.data
      }
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(`重量获取 ... ${err}`)
      return 0
    }
  }
}
