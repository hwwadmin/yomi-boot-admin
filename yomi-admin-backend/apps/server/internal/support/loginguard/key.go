package loginguard

import (
	"crypto/sha256"
	"encoding/hex"
	"net/netip"
	"strings"
)

// idLength 是 Redis key 中摘要的十六进制字符数
const idLength = 16

// accountID 返回归一化账号的摘要；空账号返回空串，表示不参与账号维度
func accountID(raw string) string {
	account := strings.ToLower(strings.TrimSpace(raw))
	if account == "" {
		return ""
	}
	return digest(account)
}

// ipID 返回客户端 IP 的摘要；IPv6 按 /64 聚合，非法 IP 归入统一标识
func ipID(raw string) string {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return digest("unknown")
	}
	addr = addr.Unmap()
	if addr.Is6() {
		if prefix, err := addr.Prefix(64); err == nil {
			addr = prefix.Masked().Addr()
		}
	}
	return digest(addr.String())
}

// digest 返回 SHA-256 前 8 字节的十六进制摘要
func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:idLength/2])
}

// ipKey 返回 IP 维度状态 key
func (g *Guard) ipKey(ip, state string) string {
	return g.prefix + ":ip:" + ip + ":" + state
}

// pairKey 返回 IP+账号维度状态 key
func (g *Guard) pairKey(account, ip, state string) string {
	return g.prefix + ":pair:" + account + ":" + ip + ":" + state
}

// accountKey 返回账号维度状态 key
func (g *Guard) accountKey(account, state string) string {
	return g.prefix + ":user:" + account + ":" + state
}
