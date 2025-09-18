package handler

import (
	"fmt"
	"speedliner-server/src/utils/structs"
	"strconv"
	"strings"
)

func formatISK(v int64) string {
	// simple 1.234.567.890 Format
	s := strconv.FormatInt(v, 10)
	n := len(s)
	if n <= 3 {
		return s
	}
	var out []byte
	pre := n % 3
	if pre > 0 {
		out = append(out, s[:pre]...)
		if n > pre {
			out = append(out, '.')
		}
	}
	for i := pre; i < n; i += 3 {
		out = append(out, s[i:i+3]...)
		if i+3 < n {
			out = append(out, '.')
		}
	}
	return string(out)
}

func buildExpressMailBody(req structs.ExpressMailRequest) string {
	b := &strings.Builder{}
	// Kopf
	fmt.Fprintln(b, "EXPRESS — PRIORITY COURIER")
	fmt.Fprintf(b, "Route: %s\n", req.Route)
	fmt.Fprintf(b, "Reward: %s ISK\n", formatISK(req.RewardISK))
	if req.CollatISK > 0 {
		fmt.Fprintf(b, "Collateral: %s ISK\n", formatISK(req.CollatISK))
	}
	fmt.Fprintf(b, "Volume: %s m³\n", formatISK(req.VolumeM3))
	fmt.Fprintln(b, "Days to complete: 1")
	fmt.Fprintln(b, "Deliver within 2–4h after acceptance.")
	if req.CustomerCharName != "" {
		fmt.Fprintf(b, "\nRequested by: %s (%d)\n", req.CustomerCharName, req.CustomerCharID)
	}
	if strings.TrimSpace(req.Notes) != "" {
		fmt.Fprintf(b, "\nNotes:\n%s\n", req.Notes)
	}
	return b.String()
}
