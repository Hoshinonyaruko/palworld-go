import { route } from 'quasar/wrappers';
import {
  createMemoryHistory,
  createRouter,
  createWebHashHistory,
  createWebHistory,
} from 'vue-router';
import routes from './routes';
import api from '../api/api';

export default route(function (/* { store, ssrContext } */) {
  const createHistory = process.env.SERVER
    ? createMemoryHistory
    : process.env.VUE_ROUTER_MODE === 'history'
    ? createWebHistory
    : createWebHashHistory;

  const Router = createRouter({
    scrollBehavior: () => ({ left: 0, top: 0 }),
    routes,
    history: createHistory(
      process.env.MODE === 'ssr' ? void 0 : process.env.VUE_ROUTER_BASE
    ),
  });

  const isLoggedIn = async () => {
    try {
      const response = await api.checkLoginStatus();
      return response.isLoggedIn;
    } catch (error) {
      console.error('Error checking login status:', error);
      return false; // 如果发生错误，则默认为未登录
    }
  };

  Router.beforeEach(async (to, from, next) => {
    const loggedIn = await isLoggedIn();
    // 测试环境下假设用户始终已登录
    //const loggedIn = true;
    if (!loggedIn) {
      // 用户尚未登录
      if (to.name !== 'login') {
        // 如果尝试访问的不是登录页，则重定向到登录页
        next({ name: 'login' });
      } else {
        // 允许访问登录页
        next();
      }
    } else {
      // 用户已登录
      if (to.name === 'login') {
        // 如果尝试访问的是登录页，但用户已经登录，则重定向到首页
        next({ path: '/index' });
      } else {
        // 允许访问其他页面
        next();
      }
    }
  });

  return Router;
});
