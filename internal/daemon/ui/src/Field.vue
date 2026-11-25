<script setup>
import { ref, useId } from "vue";
defineOptions({ inheritAttrs: false });
defineProps({ label: String, hint: String });
const model = defineModel();
const id = useId();
const message = ref("");
const edited = ref(false);
function check(e) {
  message.value = e.target.validationMessage;
}
</script>
<template>
  <div class="field">
    <label :for="id">{{ label }}</label>
    <small v-if="hint" :id="id + '-hint'">{{ hint }}</small>
    <input
      v-bind="$attrs"
      v-model="model"
      :id="id"
      :aria-invalid="Boolean(message)"
      :aria-describedby="
        [hint && id + '-hint', message && id + '-error']
          .filter(Boolean)
          .join(' ') || undefined
      "
      @input="
        edited = true;
        if ($event.target.validity.valid) message = '';
      "
      @blur="edited && check($event)"
      @invalid="check"
    />
    <small
      v-if="message"
      class="error"
      :id="id + '-error'"
      aria-live="polite"
      >{{ message }}</small
    >
  </div>
</template>
