package service

import (
	"context"
	"errors"
	pb "gocache/api/studentpb"
	dao2 "gocache/internal/pkg/student/dao"
	"gocache/utils/logger"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

/*
NewGroupManager 为groupnames 创建 Group 实例，并将它们存储在一个全局的 GroupManager 映射中
*/
func NewGroupManager(groupnames []string, currentPeerAddr string) map[string]*Group {
	// 为每个group构造一个Group实例
	for i := 0; i < len(groupnames); i++ {
		g := NewGroup(groupnames[i], "lru", 100*2*20, RetrieveFunc(func(key string) ([]byte, error) {
			start := time.Now()
			dao := dao2.NewStudentDao(context.Background())
			stus, err := dao.ShowStudentInfo(&pb.StudentRequest{
				Name:  key,
				Score: rand.Float32(),
			})

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 优化点：即使没有查询到，但是为了防止恶意攻击，这里还是往缓存中 put 一个 key 的空值并设置一个相对合理的过期时间
					return []byte{}, gorm.ErrRecordNotFound
				} else {
					return []byte{}, err
				}
			} else {
				logger.LogrusObj.Infof("成功从后端数据库中查询到学生 %s 的分数：%v", key, stus.Score)
				logger.LogrusObj.Warnf("查询数据库总耗时: %v ms", time.Since(start).Milliseconds())
			}
			return []byte(strconv.FormatFloat(stus.Score, 'f', 2, 64)), nil
		}))
		GroupManager[groupnames[i]] = g
	}
	return GroupManager
}
