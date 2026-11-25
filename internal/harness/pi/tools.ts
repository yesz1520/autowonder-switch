// This extension is loaded explicitly for one turn. Only the local bridge URL
// is passed to pi. Platform credentials stay in the daemon.
export default async function (pi) {
  const url = process.env.AUTOWONDER_PI_BRIDGE;
  async function rpc(method, params, signal?) {
    const response = await fetch(url, {
      method: "POST", redirect: "error", signal,
      headers: {"Content-Type":"application/json", Accept:"application/json, text/event-stream"},
      body: JSON.stringify({jsonrpc:"2.0", id:1, method, params})
    });
    if (!response.ok) throw new Error("Local tool bridge returned HTTP " + response.status);
    const message = await response.json();
    if (message.error) throw new Error(message.error.message);
    return message.result;
  }
  const {tools} = await rpc("tools/list", {});
  for (const tool of tools) {
    pi.registerTool({
      name: "switch_" + tool.name, label: tool.name, description: tool.description,
      parameters: tool.inputSchema,
      async execute(_id, args, signal) {
        const result = await rpc("tools/call", {name:tool.name, arguments:args}, signal);
        if (result.isError) throw new Error(result.content?.[0]?.text || "Tool call failed");
        return {content: result.content, details: {}};
      }
    });
  }
}
