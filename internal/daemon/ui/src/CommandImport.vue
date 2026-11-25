<script setup>
import { nextTick, ref } from "vue";
import { readConnectCommand } from "./command.js";

const props = defineProps({ disabled: Boolean, normalizeURL: { type: Function, required: true } });
const emit = defineEmits(["load"]);
const open = ref(false), text = ref(""), error = ref("");
const input = ref(null), trigger = ref(null);

async function toggle() {
  open.value = !open.value;
  text.value = "";
  error.value = "";
  await nextTick();
  (open.value ? input.value : trigger.value)?.focus();
}
function load() {
  try {
    const result = readConnectCommand(text.value, props.normalizeURL);
    emit("load", result);
    toggle();
  } catch (e) {
    error.value = e.message;
  }
}
</script>
<template>
  <div class="actions">
    <slot name="before" />
    <button type="button" ref="trigger" :disabled="disabled" :aria-expanded="open" @click="toggle">从指令加载</button>
    <slot name="after" />
  </div>
  <div v-if="open" class="field">
      <label for="connect-command">启动指令</label>
      <textarea id="connect-command" ref="input" v-model="text" rows="4" maxlength="65536"
        spellcheck="false" autocomplete="off" placeholder="粘贴 npx 或 PowerShell 启动指令" />
      <small v-if="error" role="alert" class="error">{{ error }}</small>
      <button type="button" :disabled="disabled || !text.trim()" @click="load">填入配置</button>
      <button type="button" @click="toggle">取消导入</button>
  </div>
</template>
