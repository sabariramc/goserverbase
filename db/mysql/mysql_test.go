package mysql_test

import (
	"context"
	"log"
	"testing"

	"github.com/sabariramc/goserverbase/db/mysql"
	"github.com/sabariramc/goserverbase/utils"
	"gotest.tools/assert"
)

type Book struct {
	BookId      string `json:"bookId" gorm:"index:Book_ak_name,unique;type:char(20);not null"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Author      string `json:"author" gorm:"type:varchar(100);not null"`
	Publication string `json:"publication" gorm:"type:varchar(100);not null"`
}

type BookModel struct {
	mysql.BaseMysqlModel
	Book
}

func (BookModel) TableName() string {
	return "book"
}

func (b *BookModel) Create(ctx context.Context, db *mysql.DB) (*BookModel, error) {
	b.SetCreateParam("Test")
	err := db.Create(b).Error
	return b, err
}

func (b *BookModel) Update(ctx context.Context, db *mysql.DB, update BookModel) (*BookModel, error) {
	b.SetUpdateParam("Test")
	err := db.Model(b).Updates(update).Error
	return b, err
}

func (b *BookModel) Delete(ctx context.Context, db *mysql.DB) error {
	return db.Delete(b).Error
}

func GetBook(ctx context.Context, db *mysql.DB, bookId string) (*BookModel, error) {
	var book BookModel
	if err := db.Where("book_id = ?", bookId).First(&book).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

func ListBook(ctx context.Context, db *mysql.DB) ([]BookModel, error) {
	var books []BookModel
	err := db.Find(&books).Error
	return books, err
}

func Migrate(db *mysql.DB) {
	err := db.AutoMigrate(&BookModel{})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMysqlConnection(t *testing.T) {
	db := mysql.NewConnection(context.Background(), MysqlTestConfig.Mysql, *MysqlTestLogger, &mysql.DebugConfig)
	Migrate(db)
	book := BookModel{Book: Book{BookId: utils.GetRandomString(10, "book")}}
	_, err := book.Create(context.TODO(), db)
	assert.NilError(t, err)
	_, err = ListBook(context.TODO(), db)
	assert.NilError(t, err)
	MysqlTestLogger.Info(context.TODO(), "Test", "Test passed")
}
