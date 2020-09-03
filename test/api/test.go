package api

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/31 5:11 下午
 * @updateDate           2020/8/31 5:11 下午
 * @version              1.0
**/
const (
	SayHello = "SayHello"
)

type TestApi interface {
	SayHello(name string) string
}
