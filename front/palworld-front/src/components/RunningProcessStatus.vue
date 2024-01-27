<template>
  <div class="text-center">
    <q-chip>
      <q-avatar color="accent" icon="memory" />
      <strong>CPU 使用率:</strong>
      <pre>{{ status?.cpu_percent.toFixed(2) }}%</pre>
    </q-chip>
    <q-chip>
      <q-avatar color="indigo" icon="sd_storage" />
      <strong>系统总内存:</strong>
      <code>{{ (status?.memory.total / 1024 ** 2).toFixed(2) }} MB</code>
    </q-chip>
    <q-chip>
      <q-avatar color="green" icon="memory" />
      <strong>可用内存:</strong>
      <code>{{ (status?.memory.available / 1024 ** 2).toFixed(2) }} MB</code>
    </q-chip>
    <q-chip>
      <q-avatar color="light-green" icon="memory" />
      <strong>内存使用率:</strong>
      <code>{{ status?.memory.percent.toFixed(2) }}%</code>
    </q-chip>
    <q-chip>
      <q-avatar color="deep-purple" icon="disc_full" />
      <strong>磁盘总容量:</strong>
      <code>{{ (status?.disk.total / 1024 ** 3).toFixed(2) }} GB</code>
    </q-chip>
    <q-chip>
      <q-avatar color="purple" icon="disc_full" />
      <strong>磁盘剩余空间:</strong>
      <code>{{ (status?.disk.free / 1024 ** 3).toFixed(2) }} GB</code>
    </q-chip>
    <q-chip>
      <q-avatar color="pink" icon="disc_full" />
      <strong>磁盘使用率:</strong>
      <code>{{ status?.disk.percent.toFixed(2) }}%</code>
    </q-chip>
  </div>
</template>
<script setup lang="ts">
const props = defineProps<{ status: any }>();
</script>
