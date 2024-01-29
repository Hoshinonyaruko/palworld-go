<template>
  <div class="my-page">
    <q-btn color="blue" label="生成机器人指令" @click="getBot"></q-btn>
    <q-btn color="primary" label="获取机器人" @click="getBotLink"></q-btn>
    <q-input v-model="apiResponse" readonly></q-input>
    <q-btn color="green" label="点击复制" @click="copyResponse"></q-btn>
    <div>如果点击复制无效,请在守护设置开启强制https或手动复制。</div>
    <div>指令包含加密后的你的面板地址,需放通webui端口到公网。</div>
    <div>建议发出给帕鲁帕鲁后立即撤回。</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { QBtn, QInput } from 'quasar';

const apiResponse = ref('');

const getBot = async () => {
  const response = await fetch('/api/getbot', {
    method: 'POST',
    credentials: 'include',
  });
  apiResponse.value = await response.text();
};

const getBotLink = async () => {
  const response = await fetch('/api/getbotlink', {
    method: 'POST',
    credentials: 'include',
  });
  apiResponse.value = await response.text();
};

const copyResponse = () => {
  navigator.clipboard.writeText(apiResponse.value).catch(() => {
    console.error('复制失败，请手动复制');
  });
};
</script>

<style lang="scss">
.my-page {
  .q-btn {
    margin-bottom: 10px;
  }

  .q-input {
    margin-bottom: 10px;
  }
}
</style>
