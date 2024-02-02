<template>
  <q-page padding>
    <q-btn icon="refresh" color="primary" @click="loadPlayers" class="q-mb-md"
      >刷新</q-btn
    >
    <q-btn icon="refresh" color="primary" @click="restartSelf" class="q-mb-md"
      >应用白名单</q-btn
    >
    <div v-if="loading">加载中...</div>
    <div v-else>
      <q-list bordered separator>
        <q-item
          v-for="player in sortedPlayers"
          :key="player.playeruid"
          clickable
          @click="selectPlayer(player)"
        >
          <q-item-section avatar>
            <q-icon
              name="online_prediction"
              v-if="player.online"
              color="green"
            />
          </q-item-section>
          <q-item-section>
            <div class="text-h6">
              {{ player.online ? '【在线】' : '' }} {{ player.name }}
            </div>
            <div>上次在线时间: {{ player.last_online }}</div>
            <div>playeruid: {{ player.playeruid }}</div>
            <div>steamid: {{ player.steamid }}</div>
          </q-item-section>
          <q-item-section side>
            <q-btn
              flat
              color="red"
              icon="block"
              @click.stop="kickOrBan(player, 'kick')"
              >踢出</q-btn
            >
            <q-btn
              flat
              color="orange"
              icon="gavel"
              @click.stop="kickOrBan(player, 'ban')"
              >封禁</q-btn
            >
            <q-btn
              flat
              color="blue"
              icon="content_copy"
              @click.stop="copyToClipboard(player.playeruid)"
              >复制UID</q-btn
            >
            <q-btn
              flat
              color="blue"
              icon="content_copy"
              @click.stop="copyToClipboard(player.steamid)"
              >复制SteamID</q-btn
            >
            <q-btn
              flat
              color="green"
              icon="content_copy"
              @click.stop="addWhite(player)"
              >加入白名单</q-btn
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
  last_online: string;
  name: string;
  online: boolean;
  playeruid: string;
  steamid: string;
}

const players = ref<Player[]>([]);
const loading = ref(true);
const $q = useQuasar();

const loadPlayers = async () => {
  loading.value = true;
  try {
    const response = await axios.get('/api/player?update=true');
    players.value = response.data;
  } catch (error) {
    console.error('API 请求失败', error);
    $q.notify({ type: 'negative', message: '加载失败' });
  } finally {
    loading.value = false;
  }
};

const restartSelf = async () => {
  loading.value = true;
  try {
    const response = await axios.get('/api/restartself', {
      withCredentials: true, // 确保携带 cookie
    });
    players.value = response.data;
  } catch (error) {
    console.error('API 重启请求发过去了', error);
    $q.notify({ type: 'positive', message: '应用白名单成功' });
  } finally {
    loading.value = false;
  }
};

onMounted(loadPlayers);

const sortedPlayers = computed(() => {
  // 使用 slice() 创建 players 数组的副本，然后进行排序
  return players.value
    .slice()
    .sort((a, b) => Number(b.online) - Number(a.online));
});

const selectPlayer = (player: Player) => {
  // 记录选中的playeruid和steamid
  console.log('选中的玩家:', player);
};

const kickOrBan = async (player: Player, type: 'kick' | 'ban') => {
  try {
    await axios.post('/api/kickorban', {
      playeruid: player.playeruid,
      steamid: player.steamid,
      type: type,
    });
    $q.notify({
      type: 'positive',
      message: `${type === 'kick' ? '踢出' : '封禁'}成功`,
    });
  } catch (error) {
    $q.notify({ type: 'negative', message: '操作失败' });
  }
};

const addWhite = async (player: Player) => {
  try {
    await axios.post('/api/addwhite', {
      playeruid: player.playeruid,
      steamid: player.steamid,
      name: player.name,
    });
    $q.notify({
      type: 'positive',
      message: `加白成功`,
    });
  } catch (error) {
    $q.notify({ type: 'negative', message: '操作失败' });
  }
};

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    $q.notify({ type: 'positive', message: '复制成功' });
  });
};
</script>

<style scoped lang="scss">
/* 可以添加自定义样式 */
</style>
