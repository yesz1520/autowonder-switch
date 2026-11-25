<script setup>
import {ref, onMounted, onUnmounted} from "vue";

const props = defineProps({identity: {type: String, required: true}});
const rows = ref(null), failed = ref(false);
const direction = {sent: "发送", received: "接收", http: "HTTP", connection: "连接"};
const outcome = {ok: "成功", failed: "失败", closed: "已关闭", retry: "重试"};
const clock = new Intl.DateTimeFormat(undefined, {dateStyle: "short", timeStyle: "medium"});
const controller = new AbortController();
let timer;
async function refresh() {
  try {
    const res = await fetch("/api/executors/" + encodeURIComponent(props.identity) + "/events", {signal: controller.signal});
    if (!res.ok) throw Error();
    rows.value = await res.json();
    failed.value = false;
  } catch {
    if (!controller.signal.aborted) failed.value = true;
  } finally {
    if (!controller.signal.aborted) timer = setTimeout(refresh, 3000);
  }
}
onMounted(refresh);
onUnmounted(() => {controller.abort(); clearTimeout(timer);});
</script>
<template>
  <p v-if="failed" role="alert">读取执行记录失败，正在重试。</p>
  <p v-if="!rows && !failed">正在读取执行记录…</p>
  <p v-else-if="rows?.length === 0" class="empty">暂无执行记录</p>
  <div v-if="rows?.length" class="event-history">
    <table>
      <caption>最近 {{ rows.length }} 条交互</caption>
      <thead><tr><th>时间</th><th>方向</th><th>事件或接口</th><th>结果</th></tr></thead>
      <tbody><tr v-for="(row, i) in rows" :key="row.time + i">
        <td><time :datetime="row.time">{{ clock.format(new Date(row.time)) }}</time></td>
        <td>{{ direction[row.direction] }}</td>
        <td><code>{{ row.operation }}</code>
          <small v-if="row.dispatchId">派单 #{{ row.dispatchId }}</small>
          <small v-if="row.conversationId">会话 #{{ row.conversationId }}</small>
          <small v-if="row.turnId">轮次 #{{ row.turnId }}</small>
          <small v-if="row.resultId">结果 {{ row.resultId }}</small>
          <small v-if="row.requestId">请求 {{ row.requestId }}</small>
        </td>
        <td>{{ row.code || outcome[row.outcome] }} <span v-if="row.status">{{ row.status }}</span> {{ row.phase }}</td>
      </tr></tbody>
    </table>
  </div>
</template>
