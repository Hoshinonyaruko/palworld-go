<template>
  <q-layout view="hHh lpR fFf">
    <!-- 顶部滑动栏 -->
    <q-header
      class="q-layout__section--marginal main-layout-header shadow-up-2 fixed-top"
    >
      <q-tabs v-model="tab" align="justify" scrollable>
        <q-tab name="guard" label="守护配置修改" />
        <q-tab name="server" label="服务端配置修改" />
        <q-tab name="engine" label="引擎配置修改" />
        <q-tab name="command" label="服务器指令" />
        <q-tab name="player-manage" label="玩家管理" />
        <q-tab name="server-check" label="主机管理" />
        <q-tab name="save-manage" label="存档管理" />
      </q-tabs>
    </q-header>

    <!-- 主页面内容区 -->
    <q-page-container class="custom-flex-fit fit column no-wrap" style="max-width: 980px;margin-left: auto;margin-right: auto;">
      <q-page padding v-if="tab === 'guard'">
        <!-- 守护配置修改页面内容 -->
        <div class="q-gutter-xs q-mt-md">
          <div class="text-subtitle2">守护配置修改</div>

          <!-- 文本输入框 -->
          <q-input
            filled
            v-model="config.processName"
            label="进程名称"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.autoLaunchWebui"
            label="启动后自动在服务器打开网页"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.communityServer"
            label="启动为社区服务器(需设置steam路径)"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.useDll"
            label="自动注入UE4SS和可输入命令控制台DLL"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.maintenanceWarningMessage"
            label="维护公告消息(英文)"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.gamePath"
            label="游戏服务端exe路径"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.gameSavePath"
            label="游戏存档路径"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.backupPath"
            label="游戏存档备份存放路径"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.steamPath"
            label="Steam安装路径(启动社区服务器用)"
            class="q-my-md"
          />

          <!-- 数字输入框 -->
          <q-input
            filled
            v-model.number="config.RestartInterval"
            type="number"
            label="定时重启服务端（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.backupInterval"
            type="number"
            label="备份间隔（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.memoryCheckInterval"
            type="number"
            label="内存占用检测时间（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.checkInterval"
            type="number"
            label="进程存活检测时间（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.memoryCleanupInterval"
            type="number"
            label="内存清理时间间隔（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.messageBroadcastInterval"
            type="number"
            label="消息广播周期（秒）"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.memoryUsageThreshold"
            type="number"
            label="服务端重启内存阈值(百分比)"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.totalMemoryGB"
            type="number"
            label="当前服务器总内存(GB)"
            class="q-my-md"
          />
          <!-- 数组类型的配置项：定期推送的消息数组 -->
          <div class="q-my-md">
            <div class="text-h6">定期推送的消息</div>
            <div
              v-for="(message, index) in config.regularMessages"
              :key="index"
              class="q-mb-sm"
            >
              <q-input
                filled
                v-model="config.regularMessages[index]"
                label="消息内容"
                dense
              />
              <q-btn
                flat
                icon="delete"
                @click="removeMessage(index)"
                class="q-ml-md"
              />
            </div>
            <q-btn
              flat
              icon="add"
              @click="addMessage"
              label="添加消息"
              class="q-mt-md"
            />
          </div>

          <!-- 数组类型的配置项：启动参数数组 -->
          <div class="q-my-md">
            <div class="text-h6">服务端启动参数</div>
            <div
              v-for="(message, index) in config.serverOptions"
              :key="index"
              class="q-mb-sm"
            >
              <q-input
                filled
                v-model="config.serverOptions[index]"
                label="启动参数"
                dense
              />
              <q-btn
                flat
                icon="delete"
                @click="removeMessageServerOptions(index)"
                class="q-ml-md"
              />
            </div>
            <q-btn
              flat
              icon="add"
              @click="addMessageServerOptions"
              label="添加参数"
              class="q-mt-md"
            />
          </div>

          <!-- 保存按钮 -->
          <q-btn
            color="primary"
            label="保存"
            @click="saveConfig"
            class="q-mt-md"
          />
          <!-- 重启服务端按钮 -->
          <q-btn
            color="secondary"
            label="重启服务端"
            @click="restartServer"
            class="q-mt-md"
          />
        </div>
      </q-page>
      <q-page padding v-if="tab === 'server'">
        <!-- 服务端配置修改页面内容 -->
        <div class="q-gutter-xs q-mt-md">
          <div class="text-subtitle2">服务端配置修改</div>
          <!-- Rcon -->
          <q-input
            filled
            v-model="config.worldSettings.adminPassword"
            label="Rcon管理员密码"
            type="password"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.rconPort"
            type="number"
            label="rcon端口号"
            class="q-my-md"
          />
          <!-- 难度和死亡掉落 -->
          <!-- 难度选择框 -->
          <div class="q-my-md">
            <q-select
              filled
              v-model="config.worldSettings.difficulty"
              :options="difficultyOptions"
              label="难度"
            />
            <q-btn
              icon="help"
              flat
              round
              dense
              @click="toggleTooltip2('difficulty')"
            />
            <q-tooltip v-if="showDifficultyTooltip">
              难度说明：简单（Eazy），困难（Difficult）
            </q-tooltip>
          </div>

          <!-- 死亡掉落选择框 -->
          <div class="q-my-md">
            <q-select
              filled
              v-model="config.worldSettings.deathPenalty"
              :options="deathPenaltyOptions"
              label="死亡掉落"
            />
            <q-btn
              icon="help"
              flat
              round
              dense
              @click="toggleTooltip('deathPenalty')"
            />
            <q-tooltip v-if="showDeathPenaltyTooltip">
              死亡掉落说明：不掉落（None），只掉落物品（Item），掉落物品和装备（ItemAndEquipment），掉落物品、装备和帕鲁（All）
            </q-tooltip>
          </div>
          <!-- 文本输入框 -->
          <q-input
            filled
            v-model="config.worldSettings.serverName"
            label="服务器名称(控制台用户名)"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.worldSettings.serverDescription"
            label="服务器描述"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.worldSettings.serverPassword"
            label="服务器进入密码"
            type="password"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.worldSettings.publicIP"
            label="公共 IP 地址"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.worldSettings.region"
            label="区域"
            class="q-my-md"
          />
          <q-input
            filled
            v-model="config.worldSettings.banListURL"
            label="封禁列表 URL"
            class="q-my-md"
          />

          <!-- 数字输入框 -->
          <q-input
            filled
            v-model.number="config.worldSettings.serverPlayerMaxNum"
            type="number"
            label="游戏服务器最大人数"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.publicPort"
            type="number"
            label="公共端口"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.dayTimeSpeedRate"
            type="number"
            label="白天时间速率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.nightTimeSpeedRate"
            type="number"
            label="夜晚时间速率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.expRate"
            type="number"
            label="经验值速率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palCaptureRate"
            type="number"
            label="Pal 捕获率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palSpawnNumRate"
            type="number"
            label="Pal 生成数量率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palDamageRateAttack"
            type="number"
            label="Pal 攻击伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palDamageRateDefense"
            type="number"
            label="Pal 防御伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerDamageRateAttack"
            type="number"
            label="玩家攻击伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerDamageRateDefense"
            type="number"
            label="玩家防御伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerStomachDecreaceRate"
            type="number"
            label="玩家饥饿下降率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerStaminaDecreaceRate"
            type="number"
            label="玩家耐力下降率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerAutoHPRegeneRate"
            type="number"
            label="玩家自动生命恢复率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.playerAutoHpRegeneRateInSleep"
            type="number"
            label="玩家睡眠中自动生命恢复率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palStomachDecreaceRate"
            type="number"
            label="Pal 饥饿下降率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palStaminaDecreaceRate"
            type="number"
            label="Pal 耐力下降率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palAutoHPRegeneRate"
            type="number"
            label="Pal 自动生命恢复率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palAutoHpRegeneRateInSleep"
            type="number"
            label="Pal 睡眠中自动生命恢复率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.buildObjectDamageRate"
            type="number"
            label="建筑物伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="
              config.worldSettings.buildObjectDeteriorationDamageRate
            "
            type="number"
            label="建筑物恶化伤害率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.collectionDropRate"
            type="number"
            label="采集掉落率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.collectionObjectHpRate"
            type="number"
            label="采集物体生命值率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="
              config.worldSettings.collectionObjectRespawnSpeedRate
            "
            type="number"
            label="采集物体重生速率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.enemyDropItemRate"
            type="number"
            label="敌人掉落物品率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.dropItemAliveMaxHours"
            type="number"
            label="掉落物品存活最大小时数"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="
              config.worldSettings.autoResetGuildTimeNoOnlinePlayers
            "
            type="number"
            label="公会自动重置无在线玩家时间"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.palEggDefaultHatchingTime"
            type="number"
            label="Pal 蛋默认孵化时间"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.workSpeedRate"
            type="number"
            label="工作速度率"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.dropItemMaxNum"
            type="number"
            label="掉落物品最大数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.dropItemMaxNum_UNKO"
            type="number"
            label="帕鲁排泄物最大数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.baseCampMaxNum"
            type="number"
            label="基地最大数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.baseCampWorkerMaxNum"
            type="number"
            label="基地工人最大数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.guildPlayerMaxNum"
            type="number"
            label="公会最大玩家数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.coopPlayerMaxNum"
            type="number"
            label="合作玩家最大数量"
            class="q-my-md"
          />
          <q-input
            filled
            v-model.number="config.worldSettings.serverPlayerMaxNum"
            type="number"
            label="服务器玩家最大数量"
            class="q-my-md"
          />

          <!-- 开关 -->
          <q-toggle
            v-model="config.worldSettings.enablePlayerToPlayerDamage"
            label="允许玩家对玩家伤害"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableFriendlyFire"
            label="允许友军火力"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableInvaderEnemy"
            label="允许侵略者敌人"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.activeUNKO"
            label="激活UNKO"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableAimAssistPad"
            label="启用手柄瞄准辅助"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableAimAssistKeyboard"
            label="启用键盘瞄准辅助"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.autoResetGuildNoOnlinePlayers"
            label="自动重置无在线玩家的公会"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableNonLoginPenalty"
            label="启用非登录处罚"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableFastTravel"
            label="启用快速旅行"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.isStartLocationSelectByMap"
            label="是否通过地图选择开始位置"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.existPlayerAfterLogout"
            label="玩家登出后保留角色"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.enableDefenseOtherGuildPlayer"
            label="启用对其他公会玩家的防御"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.isMultiplay"
            label="是否多人游戏"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.isPvP"
            label="是否开启PvP"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.canPickupOtherGuildDeathPenaltyDrop"
            label="是否能拾取其他公会死亡掉落"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.rconEnabled"
            label="启用RCON"
            class="q-my-md"
          />
          <q-toggle
            v-model="config.worldSettings.useAuth"
            label="使用认证"
            class="q-my-md"
          />

          <!-- 保存按钮 -->
          <q-btn
            color="primary"
            label="保存"
            @click="saveConfig"
            class="q-mt-md"
          />
          <!-- 重启服务端按钮 -->
          <q-btn
            color="secondary"
            label="重启服务端"
            @click="restartServer"
            class="q-mt-md"
          />
        </div>
      </q-page>
      <q-page padding v-if="tab === 'engine'">
        <!-- 引擎配置修改页面内容 -->
        <div class="q-gutter-xs q-mt-md">
          <div class="text-subtitle2">引擎配置修改</div>
          <div class="text-subtitle3">默认已为您载入最佳配置</div>
          <div class="text-subtitle4">请在了解的情况下修改</div>

          <!-- 玩家互联网速度配置 -->
          <q-input
            filled
            v-model.number="config.engine.player.ConfiguredInternetSpeed"
            type="number"
            label="玩家互联网速度 (字节/秒)"
            hint="设置假定的玩家互联网速度。高值可以减少带宽限制的可能性。"
            class="q-my-md"
          />

          <!-- 玩家局域网速度配置 -->
          <q-input
            filled
            v-model.number="config.engine.player.ConfiguredLanSpeed"
            type="number"
            label="玩家局域网速度 (字节/秒)"
            hint="设置局域网速度，确保局域网玩家可以利用最大的网络容量。"
            class="q-my-md"
          />

          <!-- 最大客户端数据传输速率 -->
          <q-input
            filled
            v-model.number="config.engine.socketsubsystemepic.MaxClientRate"
            type="number"
            label="最大客户端数据传输速率 (字节/秒)"
            hint="为所有连接设置每个客户端的最大数据传输速率，设置高值以防止数据上限。"
            class="q-my-md"
          />

          <!-- 特别针对互联网客户端的最大数据传输速率 -->
          <q-input
            filled
            v-model.number="
              config.engine.socketsubsystemepic.MaxInternetClientRate
            "
            type="number"
            label="互联网客户端最大数据传输速率 (字节/秒)"
            hint="特别针对互联网客户端，允许高容量数据传输而不受限制。"
            class="q-my-md"
          />

          <!-- 游戏引擎平滑帧率设置 -->
          <q-toggle
            v-model="config.engine.engine.bSmoothFrameRate"
            label="启用平滑帧率"
            hint="使游戏引擎平滑帧率波动，以获得更一致的游戏体验。"
            class="q-my-md"
          />

          <!-- 禁用固定帧率设置 -->
          <q-toggle
            v-model="config.engine.engine.bUseFixedFrameRate"
            label="启用动态帧率"
            hint="动态帧率，允许游戏服务端动态调整帧率以获得最佳性能。"
            class="q-my-md"
          />

          <!-- 最低可接受帧率设置 -->
          <q-input
            filled
            v-model.number="config.engine.engine.MinDesiredFrameRate"
            type="number"
            label="最低可接受帧率"
            hint="指定最低可接受帧率，确保游戏至少以这个帧率流畅运行。"
            class="q-my-md"
          />

          <!-- 平滑帧率范围的下限 -->
          <q-input
            filled
            v-model.number="
              config.engine.engine.SmoothedFrameRateRange.LowerBound.Value
            "
            type="number"
            label="平滑帧率范围下限"
            hint="设置平滑的目标帧率下限。"
            class="q-my-md"
          />

          <!-- 平滑帧率范围的上限 -->
          <q-input
            filled
            v-model.number="
              config.engine.engine.SmoothedFrameRateRange.UpperBound.Value
            "
            type="number"
            label="平滑帧率范围上限"
            hint="设置平滑的目标帧率上限。"
            class="q-my-md"
          />

          <!-- 固定帧率配置 -->
          <q-input
            filled
            v-model.number="config.engine.engine.FixedFrameRate"
            type="number"
            label="固定帧率"
            hint="设置固定帧率的值（仅当启用固定帧率时有效）。"
            class="q-my-md"
          />

          <!-- 客户端更新频率设置 -->
          <q-input
            filled
            v-model.number="config.engine.engine.NetClientTicksPerSecond"
            type="number"
            label="客户端更新频率 (次/秒)"
            hint="增加客户端的更新频率，提高响应性并减少延迟。"
            class="q-my-md"
          />
          <!-- 保存按钮 -->
          <q-btn
            color="primary"
            label="保存"
            @click="saveConfig"
            class="q-mt-md"
          />
          <!-- 重启服务端按钮 -->
          <q-btn
            color="secondary"
            label="重启服务端"
            @click="restartServer"
            class="q-mt-md"
          />
        </div>
      </q-page>
      <q-page padding v-if="tab === 'command'">
        <div class="q-pa-md">
          <div class="text-h6">服务器指令页面</div>

          <q-toggle v-model="useCommands2" label="自动填充" class="q-mb-md" />

          <q-input
            v-model="command"
            label="输入指令"
            filled
            v-on:keyup.enter="sendCommand"
          />

          <div class="q-mt-md">
            <q-card>
              <q-card-section class="bg-grey-2">
                <div
                  v-for="(message, index) in messages"
                  :key="index"
                  class="q-mb-sm"
                >
                  {{ message }}
                </div>
              </q-card-section>
            </q-card>
          </div>

          <!-- 指令快捷按钮 -->
          <div class="q-mt-md">
            <q-btn
              v-for="(cmd, index) in useCommands2 ? commands2 : commands"
              :key="index"
              :label="cmd.label"
              @click="fillCommand(cmd.prefix)"
              class="q-mr-md q-mb-md"
            />
          </div>
        </div>
      </q-page>
      <!-- 玩家管理组件 -->
      <q-page padding v-if="tab === 'player-manage'">
        <player-manage />
      </q-page>
      <q-page padding v-if="tab === 'server-check'">
        <div class="text-h6">服务器检测页面</div>
        <running-process-status
          v-if="status.process && status.process.status === 'running'"
          :status="status"
        />
      </q-page>
      <!-- 存档管理组件 -->
      <q-page padding v-if="tab === 'save-manage'">
        <save-manage />
      </q-page>

    </q-page-container>
  </q-layout>
</template>

<script setup>
import { ref, onMounted, onUnmounted, onBeforeUnmount, watch } from 'vue';
import axios from 'axios';
import { QPage, QCard, QCardSection } from 'quasar';
import RunningProcessStatus from 'components/RunningProcessStatus.vue';
import PlayerManage from 'components/PlayerManage.vue';
import SaveManage from 'components/SaveManage.vue';

//给components传递数据
const props = defineProps({
  uin: Number,
});

const status = ref(null); // 假设 ProcessInfo 是一个对象，这里使用 null 作为初始值

const tab = ref('guard'); // 默认选中守护配置修改
// 难度选项
const difficultyOptions = ['Eazy', 'None', 'Difficult'];
// 死亡掉落选项
const deathPenaltyOptions = ['None', 'Item', 'ItemAndEquipment', 'All'];

const useCommands2 = ref(false); // 切换选择框的状态

const showDifficultyTooltip = ref(false);
const showDeathPenaltyTooltip = ref(false);

const toggleTooltip = (type) => {
  showDeathPenaltyTooltip.value = !showDeathPenaltyTooltip.value;
};

const toggleTooltip2 = (type) => {
  showDifficultyTooltip.value = !showDifficultyTooltip.value;
};

// // 难度选项
// const difficultyOptions = [
//   { label: '简单', value: 'Eazy' },
//   { label: '困难', value: 'Difficult' },
// ];

// // 死亡掉落选项
// const deathPenaltyOptions = [
//   { label: '不掉落', value: 'None' },
//   { label: '只掉落物品', value: 'Item' },
//   { label: '掉落物品和装备', value: 'ItemAndEquipment' },
//   { label: '掉落物品、装备和帕鲁', value: 'All' },
// ];
const config = ref({});

// 增加一个消息到数组
const addMessage = () => {
  config.value.regularMessages.push('');
};

// 从数组中移除一个消息
const removeMessage = (index) => {
  config.value.regularMessages.splice(index, 1);
};

// 增加一个消息到数组
const addMessageServerOptions = () => {
  config.value.serverOptions.push('');
};

// 从数组中移除一个消息
const removeMessageServerOptions = (index) => {
  config.value.serverOptions.splice(index, 1);
};

onMounted(async () => {
  try {
    const response = await axios.get('/api/getjson');
    config.value = response.data;
  } catch (error) {
    console.error('Error fetching configuration:', error);
  }
});

const saveConfig = async () => {
  try {
    await axios.post('/api/savejson', config.value);
    alert('配置已保存！');
  } catch (error) {
    console.error('Error saving configuration:', error);
  }
};

const restartServer = async () => {
  try {
    const response = await axios.post(
      '/api/restart',
      {},
      {
        withCredentials: true, // 确保携带 cookie
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

// rcon终端
const command = ref('');
const messages = ref([]);
let websocket = null; // 这里声明websocket变量

const connectWebSocket = () => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
  const host = window.location.host; // 包含主机名和端口（如果有的话）
  const wsURL = `${wsProtocol}://${host}/api/ws`;

  // 使用已经声明的websocket变量
  websocket = new WebSocket(wsURL);

  websocket.onmessage = (event) => {
    messages.value.push(event.data);
  };

  websocket.onclose = () => {
    console.log('WebSocket disconnected');
    // 处理重连逻辑或通知用户
  };

  websocket.onerror = (error) => {
    console.error('WebSocket error:', error);
    // 处理错误记录或通知用户
  };
};

// 确保在组件销毁时关闭websocket
onUnmounted(() => {
  if (websocket) {
    websocket.close();
  }
});

const sendCommand = () => {
  if (websocket && websocket.readyState === WebSocket.OPEN) {
    websocket.send(command.value);
    command.value = ''; // Clear the input after sending
  } else {
    console.error('WebSocket is not connected.');
  }
};

onMounted(() => {
  connectWebSocket();
});

onUnmounted(() => {
  if (websocket) {
    websocket.close();
  }
});

// 定义指令模板
const commands2 = [
  { label: '关闭服务器', prefix: 'Shutdown {Seconds} {MessageText}' },
  { label: '强制关闭', prefix: 'DoExit' },
  { label: '广播', prefix: 'Broadcast {MessageText}' },
  { label: '踢人', prefix: 'KickPlayer {SteamID}' },
  { label: '禁止玩家进入', prefix: 'BanPlayer {SteamID}' },
  { label: '传送', prefix: 'TeleportToPlayer {SteamID}' },
  { label: '传送到自己', prefix: 'TeleportToMe {SteamID}' },
  { label: '显示玩家列表', prefix: 'ShowPlayers' },
  { label: '服务器信息', prefix: 'Info' },
  { label: '立刻存档', prefix: 'Save' },
];

// 定义指令模板
const commands = [
  { label: '关闭服务器', prefix: 'Shutdown ' },
  { label: '强制关闭', prefix: 'DoExit' },
  { label: '广播', prefix: 'Broadcast ' },
  { label: '踢人', prefix: 'KickPlayer ' },
  { label: '禁止玩家进入', prefix: 'BanPlayer ' },
  { label: '传送', prefix: 'TeleportToPlayer ' },
  { label: '传送到自己', prefix: 'TeleportToMe ' },
  { label: '显示玩家列表', prefix: 'ShowPlayers' },
  { label: '服务器信息', prefix: 'Info' },
  { label: '立刻存档', prefix: 'Save' },
];

// 填充指令到输入框的函数
const fillCommand = (cmd) => {
  command.value = cmd;
};

//服务器检测

async function updateStatus() {
  try {
    // 移除了 $q.loadingBar 的调用
    const response = await axios.get('/api/status'); // 调用 /api/status
    console.log(response.data); // 打印以检查数据结构
    status.value = response.data;
  } catch (err) {
    console.error(err);
  }
}

// 设置定时器来定期更新状态
const updateTimer = window.setInterval(() => {
  updateStatus();
}, 3000);

watch(
  () => props.uin,
  async () => {
    status.value = undefined;
    await updateStatus();
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  window.clearInterval(updateTimer);
});

void updateStatus();
</script>

<style scoped>
/* 根据需要添加CSS样式 */
</style>
