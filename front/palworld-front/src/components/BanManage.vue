<template>
  <q-page padding>
    <q-btn icon="refresh" color="primary" @click="loadPlayers" class="q-mb-md"
      >刷新</q-btn
    >

    <div v-if="loading">加载中...</div>
    <div v-else>
      <q-list bordered separator>
        <q-item
          v-for="player in players"
          :key="player.playeruid"
          clickable
          @click="selectPlayer(player)"
        >
          <q-item-section>
            <div class="text-h6">{{ player.name }}</div>
            <div>playeruid: {{ player.playeruid }}</div>
            <div>steamid: {{ player.steamid }}</div>
            <div>上次上线时间: {{ player.last_online }}</div>
            <!-- 添加显示上次上线时间 -->
          </q-item-section>
          <q-item-section side>
            <q-btn flat color="red" icon="block" @click.stop="setUnBan(player)"
              >解除封禁</q-btn
            >
          </q-item-section>
        </q-item>
      </q-list>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';

interface Player {
  name: string;
  playeruid: string;
  steamid: string;
  last_online: string;
}

const players = ref<Player[]>([]);
const loading = ref(true);
const $q = useQuasar();

const loadPlayers = async () => {
  loading.value = true;
  try {
    const response = await axios.get('/api/getban');
    // 取出bannedPlayers数组并赋值
    players.value = response.data.bannedPlayers;
  } catch (error) {
    console.error('API 请求失败', error);
    $q.notify({ type: 'negative', message: '加载失败' });
  } finally {
    loading.value = false;
  }
};

// 解除玩家封禁的函数
const setUnBan = async (player: Player) => {
  try {
    const response = await axios.post('/api/setunban', {
      steamid: player.steamid,
    });

    if (response.status === 200) {
      $q.notify({
        type: 'positive',
        message: '解除封禁成功',
      });

      // 重新加载封禁列表以反映更新
      await loadPlayers();
    }
  } catch (error) {
    console.error('API 请求失败', error);
    $q.notify({
      type: 'negative',
      message: '解除封禁失败',
    });
  }
};

onMounted(loadPlayers);

const selectPlayer = (player: Player) => {
  // 记录选中的playeruid和steamid
  console.log('选中的玩家:', player);
};
</script>

<style scoped lang="scss">
/* 可以添加自定义样式 */
</style>
