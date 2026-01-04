package core

import (
	"runtime"
)

// DecideNumWorkers は要素数とシステムリソースに基づいて最適なワーカー数を決定する
// itemCount: 処理する要素数
func DecideNumWorkers(itemCount int) int {

	// 要素数が0の場合は0を返す
	if itemCount <= 0 {
		return 0
	}

	// CPU数ベースでワーカー数を計算
	numCPU := runtime.NumCPU()
	idealWorkers := numCPU * Config.CpuMultiplier

	// 最小値と最大値の範囲内に収める
	numWorkers := max(
		Config.MinumWorkers,
		min(Config.MaximumWorkers, idealWorkers))

	// 要素数が少ない場合はワーカー数を調整
	if itemCount < numWorkers {
		// 要素数の半分程度のワーカー数にする（最小1）
		numWorkers = max(1, itemCount/2)
	}

	return numWorkers
}
