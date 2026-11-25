// Interpret the arguments as data. Never invoke a shell or expand expressions.
export function readConnectCommand(input, normalizeURL) {
  let text = input.trim();
  const wrapper = text.match(/^(?:powershell|pwsh)(?:\.exe)?\s+(?:-NoProfile\s+)?-EncodedCommand\s+["']?([A-Za-z0-9+/=\s]+?)["']?$/i);
  if (wrapper) {
    try {
      const bytes = Uint8Array.from(atob(wrapper[1]), c => c.charCodeAt(0));
      text = new TextDecoder("utf-16le", { fatal: true }).decode(bytes).trim();
    } catch {
      throw Error("PowerShell 编码无效，请重新复制完整指令。");
    }
  }
  const command = text.replace(/(?:\\|\x60)\r?\n/g, " ").match(/(?:^|;\s*)npx\s+(?:(?:-y|--yes)\s+)?(?:@alicloud\/)?autowonder(?:@[\w.-]+)?\s+connect\s+([^;]+)$/i);
  if (!command) throw Error("请粘贴完整的 AutoWonder connect 启动指令。");

  const fields = {};
  const skipped = new Set();
  let args = command[1].trim();
  while (args) {
    const flag = args.match(/^--([a-z][a-z0-9-]*)(?==|\s|$)/);
    if (!flag) throw Error("无法识别参数格式，请检查指令是否完整。");
    args = args.slice(flag[0].length).trimStart();
    const inline = args.startsWith("=");
    if (inline) args = args.slice(1);
    let value;
    if (inline || (args && !args.startsWith("--"))) {
      const arg = args.match(/^(?:"([^"]*)"|'([^']*)'|([^\s'"]+))(?=\s|$)/);
      if (!arg) throw Error("参数引号不完整，请重新复制指令。");
      value = arg[1] ?? arg[2] ?? arg[3];
      args = args.slice(arg[0].length).trimStart();
    }
    const key = flag[1];
    switch (key) {
      case "ws-url":
        fields.url = normalizeURL((value || "").replace(/\\:/g, ":"));
        if (!fields.url) throw Error("指令中的平台地址无效。");
        break;
      case "executor-id":
        if (!/^[1-9]\d{0,18}$/.test(value ?? "") || BigInt(value) > 9223372036854775807n) throw Error("执行器 ID 无效。");
        fields.executorId = value;
        break;
      case "token":
        if (!value) throw Error("指令中缺少执行器令牌。");
        fields.token = value;
        break;
      case "memory-mode":
        if (!["none", "platform", "provider-local"].includes(value)) throw Error("不支持指令中的记忆模式。");
        fields.memoryMode = value;
        break;
      case "max-tasks":
        if (!/^\d+$/.test(value ?? "") || Number(value) < 1 || Number(value) > 50) throw Error("并发容量须在 1 到 50 之间。");
        fields.capacity = Number(value);
        break;
      case "provider":
        if (["codex","claude","pi"].includes(value)) fields.harness = value;
        else skipped.add(key);
        break;
      default:
        skipped.add(key);
    }
  }
  if (!fields.url || !fields.executorId || !fields.token) throw Error("指令须包含平台地址、执行器 ID 和令牌。");
  return { fields, skipped: [...skipped] };
}
