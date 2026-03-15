import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

export default new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/chatroom.vue')
    },
    {
      path: '/discussion',
      name: 'discussion',
      component: () => import('../views/discussion.vue')
    },
    {
      path: '/discussion/:section/:postId',
      name: 'discussion-post',
      component: () => import('../views/discussionPost.vue')
    }
  ]
})
