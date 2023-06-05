// Package Coins
package Coins

import (
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	LevelArray = [...]int{0, 1, 2, 5, 10, 20, 35, 55, 75, 100, 120, 180, 260, 360, 480, 600}
)

func init() {
}

// Scoredb 分数数据库
type Scoredb gorm.DB

// Scoretable 分数结构体
type Scoretable struct {
	UID   int64 `gorm:"column:uid;primary_key"`
	Score int   `gorm:"column:score;default:0"`
}

// Signintable 签到结构体
type Signintable struct {
	UID       int64 `gorm:"column:uid;primary_key"`
	Count     int   `gorm:"column:count;default:0"`
	Coins     int   `gorm:"column:coins;default:0"`
	UpdatedAt time.Time
}

// Globaltable 总体结构体
type Globaltable struct {
	Counttime int `gorm:"column:counttime;default:0"`
	Times     string
}

// TableName ...
func (Globaltable) TableName() string {
	return "global"
}

// TableName ...
func (Scoretable) TableName() string {
	return "score"
}

// TableName ...
func (Signintable) TableName() string {
	return "sign_in"
}

// Initialize 初始化ScoreDB数据库
func Initialize(dbpath string) *Scoredb {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	gdb.AutoMigrate(&Scoretable{}).AutoMigrate(&Signintable{}).AutoMigrate(&Globaltable{})
	return (*Scoredb)(gdb)
}

// Close ...
func Close(sdb *Scoredb) error {
	db := (*gorm.DB)(sdb)
	return db.Close()
}

// GetScoreByUID 取得分数
func GetScoreByUID(sdb *Scoredb, uid int64) (s Scoretable) {
	db := (*gorm.DB)(sdb)
	db.Debug().Model(&Scoretable{}).FirstOrCreate(&s, "uid = ? ", uid)
	return s
}

// InsertOrUpdateScoreByUID 插入或更新打卡累计 初始化更新临时保存
func InsertOrUpdateScoreByUID(sdb *Scoredb, uid int64, score int) (err error) {
	db := (*gorm.DB)(sdb)
	s := Scoretable{
		UID:   uid,
		Score: score,
	}
	if err = db.Debug().Model(&Scoretable{}).First(&s, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&Scoretable{}).Create(&s).Error // newUser not user
		}
	} else {
		err = db.Debug().Model(&Scoretable{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"score": score,
			}).Error
	}
	return
}

// GetSignInByUID 取得签到次数
func GetSignInByUID(sdb *Scoredb, uid int64) (si Signintable) {
	db := (*gorm.DB)(sdb)
	db.Debug().Model(&Signintable{}).FirstOrCreate(&si, "uid = ? ", uid)
	return si
}

// GetCurrentCount 取得现在的人数
func GetCurrentCount(sdb *Scoredb, times string) (si Globaltable) {
	db := (*gorm.DB)(sdb)
	db.Debug().Model(&Globaltable{}).FirstOrCreate(&si, "times = ? ", times)
	return si
}

// InsertOrUpdateSignInCountByUID 插入或更新签到次数
func InsertOrUpdateSignInCountByUID(sdb *Scoredb, uid int64, count int) (err error) {
	db := (*gorm.DB)(sdb)
	si := Signintable{
		UID:   uid,
		Count: count,
	}
	if err = db.Debug().Model(&Signintable{}).First(&si, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Signintable{}).Create(&si) // newUser not user
		}
	} else {
		err = db.Debug().Model(&Signintable{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"count": count,
			}).Error
	}
	return
}

func InsertUserCoins(sdb *Scoredb, uid int64, coins int) (err error) { // 修改金币数值
	db := (*gorm.DB)(sdb)
	si := Signintable{
		UID:   uid,
		Coins: coins,
	}
	if err = db.Debug().Model(&Signintable{}).First(&si, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Signintable{}).Create(&si) // newUser not user
		}
	} else {
		err = db.Debug().Model(&Signintable{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"coins": coins,
			}).Error
	}
	return
}

func UpdateUserTime(sdb *Scoredb, counttime int, times string) (err error) {
	db := (*gorm.DB)(sdb)
	si := Globaltable{
		Counttime: counttime,
		Times:     times,
	}
	if err = db.Debug().Model(&Globaltable{}).First(&si, "times = ?", times).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Globaltable{}).Create(&si)
		}
	} else {
		err = db.Debug().Model(&Globaltable{}).Where("times = ?", times).Update(map[string]interface{}{
			"counttime": counttime,
		}).Error
	}
	return err
}

func CheckUserCoins(coins int) bool { // 参与一次50个柠檬片
	return coins-50 >= 0
}

func GetHourWord(t time.Time) (reply string) {
	switch {
	case 5 <= t.Hour() && t.Hour() < 12:
		reply = "今天又是新的一天呢ww"
	case 12 <= t.Hour() && t.Hour() < 14:
		reply = "吃饭了嘛w~如果没有快去吃饭哦w"
	case 14 <= t.Hour() && t.Hour() < 19:
		reply = "记得多去补点水呢~ww,辛苦了哦w"
	case 19 <= t.Hour() && t.Hour() < 24:
		reply = "晚上好吖w 今天过的开心嘛ww"
	case 0 <= t.Hour() && t.Hour() < 5:
		reply = "快去睡觉~已经很晚了w"
	default:
	}
	return
}

func GetLevel(count int) int {
	for k, v := range LevelArray {
		if count == v {
			return k
		} else if count < v {
			return k - 1
		}
	}
	return -1
}

func InsertUserCoinsForward(sdb *Scoredb, uid int64, coins int) (err error) { // 修改金币数值
	db := (*gorm.DB)(sdb)
	si := Signintable{
		UID:   uid,
		Coins: coins,
	}
	if err = db.Debug().Model(&Signintable{}).First(&si, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Signintable{}).Create(&si) // newUser not user
		}
	} else {
		err = db.Debug().Model(&Signintable{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"coins": coins,
			}).Error
	}
	return
}