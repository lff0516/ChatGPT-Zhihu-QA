package cron

import (
	"fmt"
	"qa/model"

	"github.com/robfig/cron"
)

func StartSchedule() {
	c := cron.New()

	// 每30分钟将redis数据同步到mysql
	addCronFunc(c, "@every 30m", func() {
		model.SyncUserLikeRecord()
		model.SyncAnswerLikeCount()
		model.FreeDeletedAnswersRecord()
	})

	// 每30分钟同步热榜信息
	addCronFunc(c, "@every 30m", func() {
		model.SyncHotQuestions()
	})

	c.Start()
}

func addCronFunc(c *cron.Cron, sepc string, cmd func()) {
	err := c.AddFunc(sepc, cmd)
	if err != nil {
		panic(fmt.Sprintf("定时任务异常: %v", err))
	}
}
