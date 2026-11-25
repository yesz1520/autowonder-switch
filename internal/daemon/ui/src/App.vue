<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import Field from "./Field.vue";
import {
  ArrowLeft,
  ChevronLeft,
  ChevronRight,
  ListFilter,
  Menu,
  Plus,
  RefreshCw,
  Search,
  Terminal,
  Waypoints,
  X,
} from "lucide-vue-next";
const drawer = ref(null);
const collapsed = ref(localStorage.getItem("switch.rail") !== "expanded");
function toggleRail() {
  collapsed.value = !collapsed.value;
  localStorage.setItem("switch.rail", collapsed.value ? "compact" : "expanded");
}
const narrow = matchMedia("(max-width: 768px)");
const mobile = ref(narrow.matches);
const drawerOpen = ref(false);
function resizeNav() {
  if (drawer.value?.open) drawer.value.close();
  mobile.value = narrow.matches;
}
function openDrawer() {
  drawer.value.showModal();
  drawerOpen.value = true;
}
function closeDrawer() {
  if (drawer.value?.open) drawer.value.close();
}
onMounted(() => {
  narrow.addEventListener("change", resizeNav);
});
onUnmounted(() => narrow.removeEventListener("change", resizeNav));
const harnessOptions = [{ value: "codex", label: "Codex" }];
const memoryModes = { platform: "平台记忆", "provider-local": "本机 Agent 记忆", none: "关闭记忆" };
const harnessLabel = (value) =>
  harnessOptions.find((option) => option.value === value)?.label ?? value;
const query = ref("");
const filter = ref("all");
const filters = [
  { key: "all", name: "全部" },
  { key: "connected", name: "已连接" },
  { key: "attention", name: "待连接" },
  { key: "stopped", name: "已停用" },
];
const pending = computed(
  () => items.value.filter((p) => p.enabled && p.status !== "connected").length,
);
const visible = computed(() =>
  items.value.filter((p) => {
    if (
      filter.value === "attention" &&
      (!p.enabled || p.status === "connected")
    )
      return false;
    if (
      ["connected", "stopped"].includes(filter.value) &&
      p.status !== filter.value
    )
      return false;
    const term = query.value.trim().toLowerCase();
    return [p.name, p.url, p.executorId].some((value) =>
      String(value).toLowerCase().includes(term),
    );
  }),
);
const items = ref([]),
  selected = ref(""),
  editing = ref(false),
  draft = ref({}),
  busy = ref(false);
const error = ref(""),
  notice = ref(""),
  offline = ref(false),
  ready = ref(false),
  title = ref(null);
const current = computed(() =>
  items.value.find((x) => x.key === selected.value),
);
function executorURL(value) {
  try {
    const url = new URL(value);
    if (!["ws:", "wss:", "http:", "https:"].includes(url.protocol) ||
        url.username || url.password || url.href.includes("?") || url.href.includes("#")) return "";
    url.protocol = ["wss:", "https:"].includes(url.protocol) ? "wss:" : "ws:";
    const path = url.pathname.replace(/\/+$/, "").replace(/\/(?:ws\/executor|executors)$/, "");
    url.pathname = path + "/ws/executor";
    return url.href;
  } catch {
    return "";
  }
}
const changedIdentity = computed(
  () =>
    current.value &&
    (executorURL(draft.value.url) !== current.value.url ||
      draft.value.executorId !== current.value.executorId),
);
const states = {
  connecting: "连接中",
  connected: "已连接",
  reconnecting: "等待重连",
  failed: "启动或连接失败",
  stopped: "已停用",
};
const online = computed(
  () => items.value.filter((x) => x.status === "connected").length,
);
let timer,
  generation = 0;
async function refresh() {
  const ticket = ++generation;
  try {
    const res = await fetch("/api/executors");
    if (!res.ok) throw Error();
    const rows = await res.json();
    if (ticket !== generation) return;
    items.value = rows;
    offline.value = false;
    ready.value = true;
    if (selected.value && !current.value) {
      selected.value = "";
      editing.value = false;
    }
  } catch {
    if (ticket === generation) offline.value = true;
  }
}
onMounted(() => {
  refresh();
  timer = setInterval(refresh, 3000);
});
onUnmounted(() => {
  clearInterval(timer);
  generation++;
});
async function focusTitle() {
  await nextTick();
  title.value?.focus();
}
function select(key) {
  closeDrawer();
  selected.value = key;
  editing.value = false;
  error.value = "";
  notice.value = "";
  focusTitle();
}
const reportFields = [
  { key: "reportedVersion", label: "上报版本" },
  { key: "hostname", label: "主机名" },
];
const localDefaults = ref({});
const resetting = ref({});
const defaultsFailed = ref({});
function displayedDefault(key) {
  return (
    localDefaults.value[key] ||
    (defaultsFailed.value[key] ? "读取失败" : "读取中…")
  );
}
onMounted(() => {
  for (const { key } of reportFields) readDefault(key);
});
async function readDefault(key, reset = false) {
  const target = draft.value;
  const original = target[key];
  if (reset) resetting.value[key] = true;
  try {
    const response = await fetch("/api/defaults/" + key);
    if (!response.ok) throw Error();
    const value = await response.json();
    localDefaults.value[key] = value;
    defaultsFailed.value[key] = false;
    if (
      (reset || !original) &&
      editing.value &&
      draft.value === target &&
      target[key] === original
    ) {
      target[key] = value;
      (target.restored ??= {})[key] = value;
      if (reset && key === "reportedVersion") target.reapplyVersion = true;
    }
  } catch {
    defaultsFailed.value[key] = true;
    if (reset && draft.value === target)
      error.value = "无法读取本机默认值，请检查本机配置后重试。";
  } finally {
    if (reset && draft.value === target) resetting.value[key] = false;
  }
}
function edit(profile) {
  selected.value = profile?.key || "";
  draft.value = profile
    ? { ...profile, token: "" }
    : {
        name: "",
        url: "",
        executorId: "",
        harness: "codex",
        capacity: 1,
        memoryMode: "platform",
        token: "",
        enabled: true,
      };
  if (profile?.recordedVersion)
    draft.value.reportedVersion = profile.recordedVersion;
  resetting.value = {};
  for (const { key } of reportFields) readDefault(key);
  editing.value = true;
  error.value = "";
  notice.value = "";
  focusTitle();
}
async function change(data, key, method, message) {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  notice.value = "";
  try {
    const { name, url, executorId, harness, capacity, token, enabled } = data;
    const endpoint = executorURL(url);
    if (method !== "DELETE" && !endpoint) throw Error("请填写有效的平台地址");
    const res = await fetch("/api/executors" + (key ? "/" + key : ""), {
      method,
      headers: {
        "Content-Type": "application/json",
        "X-Switch-Request": "local",
      },
      body:
        method === "DELETE"
          ? undefined
          : JSON.stringify({
              name,
              url: endpoint,
              executorId,
              harness,
              hostname:
                data.restored?.hostname === data.hostname ? "" : data.hostname,
              reportedVersion:
                data.restored?.reportedVersion === data.reportedVersion
                  ? ""
                  : data.reportedVersion,
              reapplyVersion: !!data.reapplyVersion,
              capacity: Number(capacity),
              memoryMode: data.memoryMode,
              token,
              enabled,
            }),
    });
    if (!res.ok) throw Error((await res.text()).trim());
    const result = await res.json();
    editing.value = false;
    selected.value = method === "DELETE" ? "" : result.key;
    await refresh();
    notice.value = message;
    focusTitle();
  } catch (e) {
    error.value = e.message;
  } finally {
    busy.value = false;
  }
}
function save() {
  const key = draft.value.key;
  const data = { ...draft.value };
  if (
    data.recordedVersion &&
    data.reportedVersion === data.recordedVersion &&
    !data.reapplyVersion
  )
    data.reportedVersion = current.value.reportedVersion;
  change(data, key, key ? "PUT" : "POST", "配置已保存");
}
function toggle(p) {
  if (
    p.enabled &&
    !window.confirm(
      "停用「" + p.name + "」会中断这个身份的在途任务。继续停用？",
    )
  )
    return;
  change(
    { ...p, enabled: !p.enabled },
    p.key,
    "PUT",
    p.enabled ? "执行器已停用" : "已启用执行器",
  );
}
function remove(p) {
  if (window.confirm("从本机移除「" + p.name + "」的连接配置？"))
    change({}, p.key, "DELETE", "连接配置已移除");
}
</script>
<template>
  <div class="shell" :class="{ collapsed: collapsed && !mobile }">
    <button
      v-if="mobile"
      class="mobile-toggle floating"
      aria-label="打开导航"
      :aria-expanded="drawerOpen"
      aria-controls="rail"
      @click="openDrawer"
    >
      <Menu />
    </button>
    <component
      :is="mobile ? 'dialog' : 'aside'"
      ref="drawer"
      id="rail"
      class="sidebar"
      aria-label="执行器导航"
      @close="drawerOpen = false"
      @click="
        if (
          mobile &&
          $event.target === $event.currentTarget &&
          $event.clientX > $event.currentTarget.getBoundingClientRect().right
        )
          closeDrawer();
      "
    >
      <div class="brand">
        <Waypoints /><strong class="rail-label">AutoWonder Switch</strong
        ><button
          v-if="mobile"
          autofocus
          aria-label="关闭导航"
          @click="closeDrawer"
        >
          <X />
        </button>
      </div>
      <nav aria-label="执行器导航">
        <button
          title="执行器总览"
          aria-label="执行器总览"
          :aria-current="!selected && !editing ? 'page' : undefined"
          :disabled="busy"
          @click="select('')"
        >
          <ListFilter /><span class="rail-label">执行器</span>
        </button>
        <button
          title="添加执行器"
          aria-label="新建连接"
          :aria-current="editing && !draft.key ? 'page' : undefined"
          :disabled="busy || !ready || offline"
          @click="
            closeDrawer();
            edit();
          "
        >
          <Plus /><span class="rail-label">添加执行器</span>
        </button>
      </nav>
    </component>
    <button
      v-if="!mobile"
      class="rail-toggle"
      :aria-label="collapsed ? '展开侧栏' : '收起侧栏'"
      :aria-expanded="!collapsed"
      aria-controls="rail"
      @click="toggleRail"
    >
      <component :is="collapsed ? ChevronRight : ChevronLeft" />
    </button>
    <div class="tools floating">
      <button title="刷新连接状态" aria-label="刷新连接状态" @click="refresh">
        <RefreshCw />
      </button>
    </div>
    <main>
      <div class="content">
        <p v-if="offline" class="alert" role="alert">
          维护台连接中断，正在重新连接。
        </p>
        <header>
          <div>
            <button
              v-if="editing || current"
              class="back"
              :disabled="busy"
              @click="select('')"
            >
              <ArrowLeft :size="16" />返回总览
            </button>
            <h1 ref="title" tabindex="-1">
              {{
                editing
                  ? draft.key
                    ? "编辑连接"
                    : "添加执行器"
                  : current
                    ? current.name
                    : "执行器总览"
              }}
            </h1>
            <p v-if="!editing && !current" class="meta">
              共 {{ ready ? items.length : "—" }} 个身份 <span>·</span>
              <em :class="{ idle: !online }"
                >{{ ready ? online : "—" }} 个已连接</em
              ><template v-if="pending"> · {{ pending }} 个待连接</template>
            </p>
          </div>
          <button
            v-if="!editing && !current"
            class="primary"
            :disabled="busy || !ready || offline"
            @click="edit()"
          >
            <Plus :size="17" /> 添加执行器
          </button>
        </header>
        <div role="status" class="notice">{{ notice }}</div>
        <p v-if="error" role="alert" class="alert">{{ error }}</p>
        <template v-if="editing">
          <form method="post" action="/api/executors" @submit.prevent="save">
            <fieldset :disabled="busy">
              <legend>连接配置</legend>
              <Field
                v-model="draft.name"
                label="名称"
                name="name"
                required
                maxlength="50"
                placeholder="例如：工程协作"
              />
              <div class="field">
                <label for="identity-harness">Harness</label>
                <select
                  id="identity-harness"
                  name="harness"
                  v-model="draft.harness"
                  required
                >
                  <option
                    v-for="option in harnessOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </div>
              <div class="field">
                <label for="identity-memory">记忆模式</label>
                <select id="identity-memory" name="memoryMode" v-model="draft.memoryMode" required>
                  <option v-for="(label, value) in memoryModes" :key="value" :value="value">{{ label }}</option>
                </select>
              </div>
              <div v-for="field in reportFields" :key="field.key" class="field">
                <label :for="'report-' + field.key">{{ field.label }}</label>
                <div class="report-input">
                  <input
                    :id="'report-' + field.key"
                    :name="field.key"
                    v-model="draft[field.key]"
                    :placeholder="displayedDefault(field.key)"
                    maxlength="255"
                    spellcheck="false"
                    autocapitalize="none"
                  />
                  <button
                    type="button"
                    :disabled="resetting[field.key]"
                    :aria-label="'恢复' + field.label + '默认值'"
                    :title="'恢复' + field.label + '默认值'"
                    @click="readDefault(field.key, true)"
                  >
                    <RefreshCw aria-hidden="true" />
                  </button>
                </div>
              </div>
              <Field
                v-model="draft.url"
                label="平台地址"
                name="url"
                type="url"
                required
                autocapitalize="none"
                spellcheck="false"
                placeholder="ws://autowonder.example.com/ws/executor"
              />
              <Field
                v-model="draft.executorId"
                label="执行器 ID"
                name="executorId"
                inputmode="numeric"
                pattern="[1-9][0-9]*"
                maxlength="19"
                required
              />
              <Field
                v-model="draft.token"
                label="执行器令牌"
                name="token"
                type="password"
                :required="!draft.key || !!changedIdentity"
                autocomplete="off"
                :hint="
                  changedIdentity
                    ? '更改平台或 ID 后，请填写新身份的令牌。'
                    : draft.key
                      ? '留空保留现有令牌。'
                      : ''
                "
              />
              <Field
                v-model="draft.capacity"
                label="并发容量"
                name="capacity"
                type="number"
                required
                min="1"
                max="50"
                hint="每个身份同时处理的派单和会话轮次数，范围 1–50。"
              />
              <div v-if="!draft.key" class="check">
                <input
                  id="enable"
                  name="enabled"
                  type="checkbox"
                  v-model="draft.enabled"
                /><label for="enable">保存后立即连接</label>
              </div>
            </fieldset>
            <div class="actions">
              <button class="primary" :disabled="busy || Object.values(resetting).some(Boolean)" type="submit">
                保存配置
              </button>
              <button :disabled="busy" type="button" @click="select(selected)">取消</button>
            </div>
          </form>
        </template>
        <template v-else-if="current">
          <h2>连接详情</h2>
          <dl class="card">
            <div>
              <dt>连接状态</dt>
              <dd :data-state="current.status">{{ states[current.status] }}</dd>
            </div>
            <div>
              <dt>平台地址</dt>
              <dd>{{ current.url }}</dd>
            </div>
            <div>
              <dt>执行器 ID</dt>
              <dd>{{ current.executorId }}</dd>
            </div>
            <div>
              <dt>并发容量</dt>
              <dd>{{ current.capacity }}</dd>
            </div>
            <div>
              <dt>记忆模式</dt>
              <dd>{{ memoryModes[current.memoryMode] }}</dd>
            </div>
            <div>
              <dt>主机名</dt>
              <dd>{{ current.hostname || displayedDefault("hostname") }}</dd>
            </div>
            <div>
              <dt>上报版本</dt>
              <dd>
                {{
                  current.recordedVersion ||
                  current.reportedVersion ||
                  displayedDefault("reportedVersion")
                }}
              </dd>
            </div>
            <div>
              <dt>Harness</dt>
              <dd>{{ harnessLabel(current.harness) }}</dd>
            </div>
          </dl>
          <p v-if="current.status === 'failed'" class="alert">
            请检查本机 Codex、平台地址和令牌，停用后可修改配置并重试。
          </p>
          <div class="actions">
            <button
              class="primary"
              :disabled="busy || offline"
              @click="toggle(current)"
            >
              {{ current.enabled ? "停用执行器" : "启用执行器" }}</button
            ><template v-if="!current.enabled"
              ><button :disabled="busy || offline" @click="edit(current)">
                编辑连接</button
              ><button
                class="danger"
                :disabled="busy || offline"
                @click="remove(current)"
              >
                移除
              </button></template
            >
          </div>
        </template>
        <template v-else>
          <section class="workbench" aria-label="执行器列表">
            <div v-if="items.length" class="list-tools">
              <label class="search"
                ><Search /><input
                  type="search"
                  v-model="query"
                  aria-label="搜索执行器"
                  placeholder="搜索名称、ID 或地址"
              /></label>
              <select v-model="filter" aria-label="连接状态筛选">
                <option v-for="f in filters" :value="f.key" :key="f.key">
                  {{ f.key === "all" ? "全部状态" : f.name }}
                </option>
              </select>
            </div>
            <div v-if="!ready" class="empty">正在读取本机配置…</div>
            <div v-else-if="!items.length" class="empty">
              <span class="terminal"><Terminal /></span>
              <h3>还没有执行器</h3>
              <p>添加第一个身份，将本机 Codex 连接到平台。</p>
            </div>
            <div v-else-if="!visible.length" class="empty">
              <Search />
              <h3>没有匹配的执行器</h3>
              <button
                @click="
                  query = '';
                  filter = 'all';
                "
              >
                清除筛选
              </button>
            </div>
            <ul v-if="visible.length" class="identity-list">
              <li v-for="p in visible" :key="p.key">
                <button
                  class="identity-row"
                  :disabled="busy"
                  @click="select(p.key)"
                  :aria-label="'查看' + p.name + '的连接详情'"
                >
                  <span class="identity-info"
                    ><strong>{{ p.name }}</strong
                    ><small
                      >{{ p.url }} <span>· #{{ p.executorId }}</span></small
                    ></span
                  >
                  <span class="status" :data-state="p.status">{{
                    states[p.status]
                  }}</span
                  ><ChevronRight />
                </button>
              </li>
            </ul></section
        ></template>
      </div>
    </main>
  </div>
</template>
