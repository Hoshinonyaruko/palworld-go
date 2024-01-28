<template>
  <q-page padding>
    <div class="q-mb-md">
      <q-btn label="刷新" color="primary" @click="fetchSaveList" />
      <q-btn label="立即保存" color="primary" @click="saveNow" />
    </div>

    <q-list bordered>
      <q-item v-for="save in saveList" :key="save" clickable>
        <template v-slot:prepend>
          <q-checkbox v-model="selectedSaves" :value="save" />
        </template>
        <q-item-section>{{ save }}</q-item-section>
        <q-item-section side>
          <q-btn flat @click="confirmRestore(save)"
            ><q-icon name="restore" /> 回档到此刻</q-btn
          >
        </q-item-section>
      </q-item>
    </q-list>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import {
  QPage,
  QList,
  QItem,
  QCheckbox,
  QItemSection,
  QBtn,
  QIcon,
} from 'quasar';
import axios from 'axios';
import { useQuasar } from 'quasar';

const $q = useQuasar();
const saveList = ref<string[]>([]);
const selectedSaves = ref<string[]>([]);

const fetchSaveList = async () => {
  try {
    const response = await axios.get('/api/getsavelist');
    saveList.value = response.data;
  } catch (error) {
    console.error(error);
  }
};

const confirmRestore = (saveName: string) => {
  $q.dialog({
    title: '确认',
    message: `您确定要回档到 "${saveName}" 吗？`,
    cancel: true,
    persistent: true,
  }).onOk(() => restoreSave(saveName));
};

const restoreSave = async (saveName: string) => {
  try {
    const response = await axios.post('/api/changesave', { path: saveName });
    if (response.status === 200) {
      $q.notify({
        color: 'green',
        textColor: 'white',
        icon: 'cloud_done',
        message: '回档成功',
      });
    }
  } catch (error) {
    console.error(error);
    $q.notify({
      color: 'red',
      textColor: 'white',
      icon: 'error',
      message: '回档失败',
    });
  }
};

const saveNow = async () => {
  // 获取当前时间的 Unix 时间戳（秒）
  const timestamp = Math.floor(Date.now() / 1000);

  try {
    await axios.post('/api/savenow', { timestamp }, { withCredentials: true });
  } catch (error) {
    console.error(error);
  }
};

const deleteSaves = async () => {
  try {
    await axios.post('/api/delsave', { saves: selectedSaves.value });
  } catch (error) {
    console.error(error);
  }
};

onMounted(fetchSaveList);
</script>

<style scoped lang="scss">
// Custom styles here
</style>
