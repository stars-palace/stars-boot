package task

import (
	"time"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 * 后台任务
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/31
 * @time: 22:28
 * @description:
 */
//定义一个任务函数
type Job func() error
type BackgroundTask struct {
	Time1   time.Duration
	running bool
	jobs    []Job
}

// Run 运行后台任务
func (b *BackgroundTask) Run() error {
	if b.running {
		return nil
	}
	b.running = true
	//运行
	b.run()
	return nil
}

func (b *BackgroundTask) run() {
	ticker := time.NewTicker(b.Time1)
	go func() {
		for _ = range ticker.C {
			//判读任务是否开启
			if b.running {
				for _, f := range b.jobs {
					//执行任务
					f()
				}
			}
		}
	}()
}

// Stop 停止后台任务
func (b *BackgroundTask) Stop() error {
	if !b.running {
		return nil
	}
	b.running = false
	return nil
}

// AddJob 添加工作
func (b *BackgroundTask) AddJob(job Job) {
	b.jobs = append(b.jobs, job)
}
