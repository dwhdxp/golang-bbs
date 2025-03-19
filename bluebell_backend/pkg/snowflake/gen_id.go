package snowflake

import (
	"fmt"
	"github.com/sony/sonyflake"
	"time"
)

var (
	sonyFlake     *sonyflake.Sonyflake // 实例
	sonyMachineID uint16               // 机器ID
)

func getMachineID() (uint16, error) { // 返回全局定义的机器ID
	return sonyMachineID, nil
}

// Init 需传入当前的机器ID
func Init(startTime string, machineId uint16) (err error) {
	sonyMachineID = machineId
	t, _ := time.Parse("2006-01-02", startTime) // 初始化时间戳
	settings := sonyflake.Settings{             // 生成全局配置
		StartTime: t,
		MachineID: getMachineID,
	}
	// 基于配置生成sonyflake节点
	sonyFlake = sonyflake.NewSonyflake(settings)
	return
}

// GetID 生成分布式ID
func GetID() (id uint64, err error) { // 拿到sonyFlake节点生成id值
	if sonyFlake == nil {
		err = fmt.Errorf("snoy flake not inited")
		return
	}

	id, err = sonyFlake.NextID()
	return
}
