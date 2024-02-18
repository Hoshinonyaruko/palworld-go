<template>
  <q-page class="row justify-center">
    <q-card
      class="col-12 col-xs-8 col-sm-6 col-md-4 shadow q-pa-md self-center"
    >
      <q-card-section>
        <div class="text-h5"><q-icon name="login" color="accent" /> 登录</div>
      </q-card-section>
      <q-separator />
      <q-form
        autocorrect="off"
        autocapitalize="off"
        autocomplete="off"
        spellcheck="false"
        @submit.prevent="login"
        @reset="clearForm"
      >
        <q-card-section class="q-gutter-md">
          <q-input v-model="username" filled clearable label="用户名" required>
            <template v-slot:prepend><q-icon name="person" /></template>
          </q-input>

          <q-input
            v-model="password"
            type="password"
            filled
            clearable
            label="密码"
            required
          >
            <template v-slot:prepend><q-icon name="lock" /></template>
          </q-input>
        </q-card-section>
        <q-separator />
        <q-card-actions class="justify-center">
          <q-btn flat color="positive" type="submit" icon="check">登录</q-btn>
          <q-btn flat color="negative" type="reset" icon="clear">清除</q-btn>
        </q-card-actions>
      </q-form>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import api from '../api/api';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';

const $router = useRouter();
const isLoggedIn = ref(false);
const username = ref('');
const password = ref('');
const loginError = ref('');
const $q = useQuasar();

async function checkLoggedIn() {
  try {
    // Await the axios promise, then the function it resolves to, then destructure the data property from the result
    // 直接从 api.checkLoginStatus() 获取返回值
    const loginStatus = await api.checkLoginStatus();
    isLoggedIn.value = loginStatus.isLoggedIn;
    if (isLoggedIn.value) {
      void $router.push('/index');
    }
  } catch (err) {
    console.error('Error checking login status:', err);
    isLoggedIn.value = false;
  }
}

function clearForm() {
  username.value = '';
  password.value = '';
}

async function login() {
  if (!username.value || !password.value) return;
  try {
    // 直接获取 loginApi 方法的返回值
    const loginResponse = await api.loginApi(username.value, password.value);
    if (loginResponse.isLoggedIn) {
      isLoggedIn.value = true;
      void $router.push('/index');
    } else {
      loginError.value =
      'Login failed, please check the username and password. Please refer to the program\'s command line window for the default username and password.\n登录失败，请检查用户名和密码。请查看程序命令行窗口输出的默认用户名密码。';
      // 显示通知
      $q.notify({
        color: 'negative',
        position: 'top',
        message: loginError.value,
        icon: 'report_problem',
      });
    }
  } catch (err) {
    loginError.value =
    'Login failed, please check the username and password. Please refer to the program\'s command line window for the default username and password.\n登录失败，请检查用户名和密码。请查看程序命令行窗口输出的默认用户名密码。';
    $q.notify({
      color: 'negative',
      position: 'top',
      message: loginError.value,
      icon: 'report_problem',
    });
  }
}

onMounted(() => {
  checkLoggedIn().catch((error) => {
    console.error('Failed to check login status:', error);
  });
});
</script>

<style scoped>
/* 如果需要添加或修改样式，可以在这里进行 */
</style>
