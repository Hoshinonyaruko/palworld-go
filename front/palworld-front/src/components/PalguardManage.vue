<template>
  <q-page class="q-pa-md">
    <div class="q-mb-md">
      <q-btn label="保存配置" color="primary" @click="saveConfig" />
      <q-btn
        label="重启服务器"
        color="negative"
        @click="restartServer"
        class="q-ml-md"
      />
    </div>
    <q-form class="q-gutter-md">
      <q-toggle v-model="config.RCONbase64" label="RCON 基于 Base64" />
      <q-field label="管理员 IP">
        <q-chip
          v-for="(ip, index) in config.adminIPs"
          :key="index"
          removable
          @remove="config.adminIPs.splice(index, 1)"
        >
          {{ ip }}
        </q-chip>
        <q-input
          v-model="newAdminIP"
          dense
          placeholder="添加新 IP"
          @keyup.enter="addAdminIP"
        />
      </q-field>
      <q-toggle v-model="config.allowAdminCheats" label="允许管理员作弊" />
      <q-toggle
        v-model="config.allowNoSpoilModWhichCanCrashYourServer"
        label="允许NoSpoil模组（可能会崩溃服务器）"
      />
      <q-toggle
        v-model="config.allowSpawningNPCItems"
        label="允许生成 NPC 物品"
      />
      <q-toggle v-model="config.announceConnections" label="公告连接" />
      <q-toggle v-model="config.announcePunishments" label="公告惩罚" />
      <q-field label="禁止的聊天词">
        <q-chip
          v-for="(word, index) in config.bannedChatWords"
          :key="index"
          removable
          @remove="config.bannedChatWords.splice(index, 1)"
        >
          {{ word }}
        </q-chip>
        <q-input
          v-model="newBannedChatWord"
          dense
          placeholder="添加新聊天禁词"
          @keyup.enter="addBannedChatWord"
        />
      </q-field>
      <q-field label="禁止的玩家昵称">
        <q-chip
          v-for="(name, index) in config.bannedNames"
          :key="index"
          removable
          @remove="config.bannedNames.splice(index, 1)"
        >
          {{ name }}
        </q-chip>
        <q-input
          v-model="newBannedName"
          dense
          placeholder="添加新禁止玩家昵称"
          @keyup.enter="addBannedName"
        />
      </q-field>
      <q-toggle
        v-model="config.blockNPCVendorCapture"
        label="阻止 NPC 商人捕获"
      />
      <q-toggle
        v-model="config.blockTowerBossCapture"
        label="阻止塔楼 Boss 捕获"
      />
      <q-toggle v-model="config.chatBypassWait" label="聊天绕过等待" />
      <q-toggle v-model="config.isChineseCmd" label="中文命令" />
      <q-toggle v-model="config.isStealingAllowed" label="允许偷窃" />
      <q-toggle v-model="config.logChat" label="输出聊天信息日志" />
      <q-toggle v-model="config.logNetworking" label="输出网络日志" />
      <q-toggle v-model="config.logRCON" label="输出RCON日志" />
      <q-input
        v-model="config.pveMaxToPalBanThreshold"
        label="PvE 最大宠物封禁阈值"
        type="number"
      />
      <q-input
        v-model="config.pvpMaxToBuildingDamage"
        label="PvP 对建筑伤害最大值"
        type="number"
      />
      <q-input
        v-model="config.pvpMaxToPalDamage"
        label="PvP 对宠物伤害最大值"
        type="number"
      />
      <q-input
        v-model="config.pvpMaxToPlayerDamage"
        label="PvP 对玩家伤害最大值"
        type="number"
      />
      <q-toggle v-model="config.shouldBanCheaters" label="应封禁作弊者" />
      <q-toggle v-model="config.shouldIPBanCheaters" label="应 IP 封禁作弊者" />
      <q-toggle v-model="config.shouldKickCheaters" label="应踢出作弊者" />
      <q-toggle v-model="config.shouldWarnCheaters" label="应警告作弊者" />
      <q-toggle
        v-model="config.shouldWarnCheatersReason"
        label="警告作弊者原因"
      />
      <q-toggle v-model="config.steamidProtection" label="Steam ID 保护" />
      <q-toggle v-model="config.useAdminWhitelist" label="使用管理员白名单" />
      <q-toggle v-model="config.useWhitelist" label="使用白名单" />
      <q-input
        v-model="config.whitelistMessage"
        label="白名单消息"
        type="textarea"
      />
    </q-form>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
const $q = useQuasar();
const config = ref({});

onMounted(async () => {
  try {
    const response = await axios.get('/api/getpalguardjson', {
      withCredentials: true,
    });
    config.value = response.data;
    $q.notify({
      type: 'positive',
      message: '配置加载成功',
    });
  } catch (error) {
    console.error('Error fetching configuration:', error);
    $q.notify({
      type: 'negative',
      message: '获取配置失败',
    });
  }
});

const saveConfig = async () => {
  try {
    await axios.post('/api/savepalguardjson', config.value, {
      withCredentials: true,
    });
    $q.notify({
      type: 'positive',
      message: '配置已保存！',
    });
  } catch (error) {
    console.error('Error saving configuration:', error);
    $q.notify({
      type: 'negative',
      message: '保存配置失败',
    });
  }
};

const restartServer = async () => {
  try {
    const response = await axios.post(
      '/api/restart',
      {},
      {
        withCredentials: true,
      }
    );
    if (response.status === 200) {
      alert('服务器重启命令已发送！');
    } else {
      console.error('服务器重启失败：', response.status);
    }
  } catch (error) {
    console.error('Error sending restart command:', error);
  }
};
</script>
