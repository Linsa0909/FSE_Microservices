package configclient

import "log"

// PrintDiagnostics 打印配置加载诊断报告
// knownKeys: 本服务认识的所有配置 key
func PrintDiagnostics(serviceName string, raw map[string]string, knownKeys []string) {
	log.Printf("[Diagnostics] ── %s 配置加载诊断 ──", serviceName)
	log.Printf("  配置中心返回 %d 个 key", len(raw))
	for k, v := range raw {
		log.Printf("    %s = %s", k, v)
	}

	// 检查未被识别的多余 key
	known := make(map[string]bool)
	for _, k := range knownKeys {
		known[k] = true
	}
	for k := range raw {
		if !known[k] {
			log.Printf("  ⚠️ 未知配置项 %q — %s 未使用此配置", k, serviceName)
		}
	}
	log.Printf("[Diagnostics] ────────────────────────────")
}
