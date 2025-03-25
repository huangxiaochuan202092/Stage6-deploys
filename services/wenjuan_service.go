package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"proapp/config"
	"proapp/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// 修改创建问卷函数
func CreateWenjuan(wenjuan *models.Wenjuan) error {
	// 开启事务
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 验证问题格式
		var questions []string
		if err := json.Unmarshal([]byte(wenjuan.Content), &questions); err != nil {
			return fmt.Errorf("问题格式无效: %w", err)
		}

		if len(questions) == 0 {
			return errors.New("至少需要一个问题")
		}

		// 设置默认值
		if wenjuan.Status == "" {
			wenjuan.Status = "draft"
		}

		// 创建问卷
		if err := tx.Create(wenjuan).Error; err != nil {
			return fmt.Errorf("创建问卷失败: %w", err)
		}

		return nil
	})
}

// 创建问卷并关联分类
func CreateWenjuanWithCategory(wenjuan *models.Wenjuan, categoryId uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 验证问题格式
		var questions []string
		if err := json.Unmarshal([]byte(wenjuan.Content), &questions); err != nil {
			return fmt.Errorf("问题格式无效: %w", err)
		}

		if len(questions) == 0 {
			return errors.New("至少需要一个问题")
		}

		// 设置默认值
		if wenjuan.Status == "" {
			wenjuan.Status = "draft"
		}

		// 创建问卷
		if err := tx.Create(wenjuan).Error; err != nil {
			return fmt.Errorf("创建问卷失败: %w", err)
		}

		// 如果指定了分类，创建关联
		if categoryId > 0 {
			relation := models.WenjuanCategory{
				WenjuanID:  wenjuan.ID,
				CategoryID: categoryId,
			}
			if err := tx.Create(&relation).Error; err != nil {
				return fmt.Errorf("关联分类失败: %w", err)
			}
		}

		return nil
	})
}

// 修改获取问卷列表函数
func GetAllWenjuans(page int, pageSize int, isPinned *bool) (gin.H, error) {
	var wenjuans []models.Wenjuan
	var total int64

	query := config.DB.Model(&models.Wenjuan{})

	// 添加排序条件：先按置顶状态降序，再按创建时间降序
	query = query.Order("is_pinned DESC, created_at DESC")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&wenjuans).Error; err != nil {
		return nil, err
	}

	// 处理返回数据格式
	var list []gin.H
	for _, w := range wenjuans {
		item := gin.H{
			"id":         w.ID,
			"title":      w.Title,
			"content":    w.Content,
			"status":     w.Status,
			"created_at": w.CreatedAt.Format("2006-01-02 15:04:05"),
			"deadline":   w.Deadline,
			"is_pinned":  w.IsPinned, // 确保返回置顶状态
			"answers":    w.Answers,
		}
		list = append(list, item)
	}

	return gin.H{
		"total": total,
		"list":  list,
		"page":  page,
		"size":  pageSize,
	}, nil
}

// 获取问卷答案
func GetWenjuanAnswer(id int, answerId int) (*models.WenjuanAnswer, error) {
	var answer models.WenjuanAnswer
	if err := config.DB.Where("id = ? AND wenjuan_id = ?", answerId, id).First(&answer).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

// 修改获取问卷详情的函数
func GetWenjuanById(id int) (*models.Wenjuan, error) {
	var wenjuan models.Wenjuan
	result := config.DB.
		Preload("Answers").
		First(&wenjuan, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("问卷不存在")
		}
		return nil, result.Error
	}

	// 验证并规范化 Content 格式
	var questions []string
	if err := json.Unmarshal([]byte(wenjuan.Content), &questions); err != nil {
		// 如果解析失败，将 Content 作为单个问题
		questions = []string{wenjuan.Content}
		// 更新为规范格式
		if content, err := json.Marshal(questions); err == nil {
			wenjuan.Content = string(content)
			// 更新数据库中的格式
			config.DB.Model(&wenjuan).Update("content", string(content))
		}
	}

	// 确保其他字段有默认值
	if wenjuan.Title == "" {
		wenjuan.Title = "无标题"
	}
	if wenjuan.Status == "" {
		wenjuan.Status = "draft"
	}
	if wenjuan.Answers == nil {
		wenjuan.Answers = []models.WenjuanAnswer{}
	}

	return &wenjuan, nil
}

// 修改更新问卷函数
func UpdateWenjuan(id int, updates map[string]interface{}) error {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var wenjuan models.Wenjuan
	if err := tx.First(&wenjuan, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("问卷不存在: %w", err)
	}

	// 更新基本字段
	if title, ok := updates["title"].(string); ok && title != "" {
		wenjuan.Title = title
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		wenjuan.Status = status
	}

	// 处理 content 字段
	if content, ok := updates["content"].(string); ok && content != "" {
		var questions []string
		if err := json.Unmarshal([]byte(content), &questions); err != nil {
			tx.Rollback()
			return fmt.Errorf("问题格式无效: %w", err)
		}
		if len(questions) == 0 {
			tx.Rollback()
			return errors.New("问题不能为空")
		}
		wenjuan.Content = content
	}

	// 保存更新
	if err := tx.Save(&wenjuan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存失败: %w", err)
	}

	return tx.Commit().Error
}

// 修改删除问卷函数
func DeleteWenjuan(id int) error {
	tx := config.DB.Begin() // 开启事务

	// Step 1: 删除关联的分类关系
	if err := tx.Where("wenjuan_id = ?", id).Delete(&models.WenjuanCategory{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除分类关系失败: %w", err)
	}

	// Step 2: 删除关联的答案
	if err := tx.Where("wenjuan_id = ?", id).Delete(&models.WenjuanAnswer{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除答案失败: %w", err)
	}

	// Step 3: 删除问卷本身
	if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Wenjuan{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除问卷失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 验证删除是否成功
	var count int64
	if err := config.DB.Unscoped().Model(&models.Wenjuan{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("删除失败，记录仍然存在")
	}

	return nil
}

// 修改提交答案函数
func SubmitWenjuanAnswer(wenjuanId int, answer string) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 检查问卷是否存在
		var wenjuan models.Wenjuan
		if err := tx.First(&wenjuan, wenjuanId).Error; err != nil {
			return err
		}

		// 创建答案记录
		wenjuanAnswer := models.WenjuanAnswer{
			WenjuanID: uint(wenjuanId),
			Answer:    answer,
		}

		return tx.Create(&wenjuanAnswer).Error
	})
}

// 修改更新答案的函数
func UpdateWenjuanAnswer(wenjuanId int, answerId int, newAnswer string) (*models.WenjuanAnswer, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先检查问卷是否存在
	var wenjuan models.Wenjuan
	if err := tx.First(&wenjuan, wenjuanId).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("问卷不存在: %w", err)
	}

	// 验证新答案格式是否为有效的JSON数组
	var answers []string
	if err := json.Unmarshal([]byte(newAnswer), &answers); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("答案格式无效，应为JSON数组: %w", err)
	}

	// 获取问题列表
	var questions []string
	if err := json.Unmarshal([]byte(wenjuan.Content), &questions); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("问卷问题格式无效: %w", err)
	}

	// 验证答案数量是否匹配问题数量
	if len(answers) != len(questions) {
		tx.Rollback()
		return nil, fmt.Errorf("答案数量(%d)与问题数量(%d)不匹配", len(answers), len(questions))
	}

	// 查找或创建答案记录
	answer := &models.WenjuanAnswer{}
	if answerId > 0 {
		// 如果提供了answerId，查找现有答案
		if err := tx.First(answer, answerId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 如果答案不存在，创建新答案
				answer.WenjuanID = uint(wenjuanId)
				answer.Answer = newAnswer
				if err := tx.Create(answer).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("创建新答案失败: %w", err)
				}
			} else {
				tx.Rollback()
				return nil, fmt.Errorf("查询答案失败: %w", err)
			}
		} else {
			// 更新现有答案
			answer.Answer = newAnswer
			if err := tx.Save(answer).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("更新答案失败: %w", err)
			}
		}
	} else {
		// 如果没有提供answerId，直接创建新答案
		answer.WenjuanID = uint(wenjuanId)
		answer.Answer = newAnswer
		if err := tx.Create(answer).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建答案失败: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return answer, nil
}

// 修改删除答案的函数
func DeleteWenjuanAnswer(wenjuanId int, answerId int) error {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查答案是否存在且属于指定问卷
	var answer models.WenjuanAnswer
	if err := tx.Where("id = ? AND wenjuan_id = ?", answerId, wenjuanId).First(&answer).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("答案不存在或不属于指定问卷: %w", err)
	}

	// 直接删除答案
	if err := tx.Delete(&answer).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除答案失败: %w", err)
	}

	return tx.Commit().Error
}

// 修改置顶问卷函数
func PinWenjuan(wenjuanId int) error {
	return config.DB.Model(&models.Wenjuan{}).
		Where("id = ?", wenjuanId).
		Update("is_pinned", true).Error
}

// 修改取消置顶函数
func UnpinWenjuan(wenjuanId int) error {
	return config.DB.Model(&models.Wenjuan{}).
		Where("id = ?", wenjuanId).
		Update("is_pinned", false).Error
}

// GetAllCategories 获取所有分类列表
func GetAllCategories() (gin.H, error) {
	var categories []models.Category

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 使用左连接获取分类及其关联的问卷数量
		if err := tx.Model(&models.Category{}).
			Select("categories.*, COUNT(DISTINCT wenjuan_categories.wenjuan_id) as wenjuan_count").
			Joins("LEFT JOIN wenjuan_categories ON categories.id = wenjuan_categories.category_id").
			Group("categories.id").
			Find(&categories).Error; err != nil {
			return fmt.Errorf("查询分类失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 格式化返回数据
	list := make([]gin.H, len(categories))
	for i, cat := range categories {
		list[i] = gin.H{
			"id":            cat.ID,
			"name":          cat.Name,
			"description":   cat.Description,
			"created_at":    cat.CreatedAt.Format("2006-01-02 15:04:05"),
			"wenjuan_count": 0, // 默认值
		}
	}

	return gin.H{
		"code": 200,
		"data": gin.H{
			"total": len(categories),
			"list":  list,
		},
	}, nil
}

// GetCategoryById 获取分类详情
func GetCategoryById(id int) (*models.Category, error) {
	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		return nil, fmt.Errorf("分类不存在: %w", err)
	}
	return &category, nil
}

func CategoryExistsByName(name string) (bool, error) {
	var count int64
	err := config.DB.Model(&models.Category{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCategory 创建新分类
func CreateCategory(name, description string) error {
	category := models.Category{
		Name:        name,
		Description: description,
	}

	return config.DB.Create(&category).Error
}

// 更新分类
func UpdateCategory(id int, updates map[string]interface{}) error {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询分类
	var category models.Category
	if err := tx.First(&category, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("分类不存在: %w", err)
	}

	// 更新分类名称
	if name, ok := updates["name"].(string); ok && name != "" {
		category.Name = name
	}

	// 更新分类描述
	if desc, ok := updates["description"].(string); ok {
		category.Description = desc
	}

	// 保存更新
	if err := tx.Save(&category).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存失败: %w", err)
	}

	return tx.Commit().Error
}

// 删除分类
func DeleteCategory(id int) error {
	return config.DB.Where("id = ?", id).Delete(&models.Category{}).Error
}

// 添加分类到问卷
func AddCategoryToWenjuan(wenjuanId, categoryId int) error {
	// 插入中间表记录
	relation := models.WenjuanCategory{
		WenjuanID:  uint(wenjuanId),
		CategoryID: uint(categoryId),
	}
	if err := config.DB.Create(&relation).Error; err != nil {
		return fmt.Errorf("关联失败: %w", err)
	}
	return nil
}

// 下载问卷和答案
// ExportWenjuanAsPDF 将问卷及答案导出为 PDF 格式
func DownloadWenjuanAndAnswers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	pdfData, err := ExportWenjuanAsPDF(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=wenjuan_%d.pdf", id))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// ExportWenjuanAsPDF 将问卷及答案导出为 PDF 格式
func ExportWenjuanAsPDF(id uint) ([]byte, error) {
	var wenjuan models.Wenjuan

	// 查找问卷及其答案
	if err := config.DB.Preload("Answers").First(&wenjuan, id).Error; err != nil {
		log.Printf("Database query failed: %v", err)
		return nil, errors.New("问卷不存在")
	}

	// 初始化 PDF 文档
	pdf := gofpdf.New("P", "mm", "A4", "")

	// 使用默认字体（Arial）
	pdf.SetFont("Arial", "B", 16)

	// 添加标题
	pdf.AddPage()
	pdf.CellFormat(40, 10, "Questionnaire Details", "", 1, "L", false, 0, "")

	// 添加问卷基本信息
	pdf.SetFont("Arial", "", 12)
	safeTitle := strings.TrimSpace(wenjuan.Title)
	if safeTitle == "" {
		safeTitle = "Untitled"
	}
	pdf.CellFormat(40, 10, fmt.Sprintf("Title: %s", safeTitle), "", 1, "L", false, 0, "")

	safeContent := strings.TrimSpace(wenjuan.Content)
	if safeContent == "" {
		safeContent = "No Content"
	}
	pdf.CellFormat(40, 10, fmt.Sprintf("Content: %s", safeContent), "", 1, "L", false, 0, "")

	pdf.CellFormat(40, 10, fmt.Sprintf("Status: %s", wenjuan.Status), "", 1, "L", false, 0, "")
	pdf.Ln(15)

	// 添加答案列表标题
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(40, 10, "Answer List", "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// 遍历答案并添加到 PDF
	pdf.SetFont("Arial", "", 12)
	for _, answer := range wenjuan.Answers {
		safeAnswer := strings.TrimSpace(answer.Answer)
		if safeAnswer == "" {
			safeAnswer = "No Answer"
		}
		pdf.CellFormat(40, 10, fmt.Sprintf("ID: %d", answer.ID), "", 1, "L", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("Answer: %s", safeAnswer), "", 1, "L", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("Created At: %s", answer.CreatedAt.Format(time.RFC3339)), "", 1, "L", false, 0, "")
		pdf.Ln(10)
	}

	// 输出 PDF 数据到缓冲区
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Printf("Failed to generate PDF: %v", err)
		return nil, errors.New("PDF 生成失败：" + err.Error())
	}

	return buf.Bytes(), nil
}

// SearchWenjuanByTitle 根据标题搜索问卷
func SearchWenjuanByTitle(title string, page, pageSize int) (gin.H, error) {
	var wenjuans []models.Wenjuan
	var total int64

	// 构建查询
	query := config.DB.Model(&models.Wenjuan{})

	// 添加标题搜索条件
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// 验证页码范围
	if page < 1 {
		page = 1
	}
	if int64(page) > totalPages {
		page = int(totalPages)
	}
	if page < 1 {
		page = 1
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&wenjuans).Error

	if err != nil {
		return nil, err
	}

	// 加载关联数据
	for i := range wenjuans {
		config.DB.Model(&wenjuans[i]).Association("Answers").Find(&wenjuans[i].Answers)
	}

	// 构建返回数据
	var list []gin.H
	for _, w := range wenjuans {
		item := gin.H{
			"id":         w.ID,
			"title":      w.Title,
			"status":     w.Status,
			"created_at": w.CreatedAt.Format("2006-01-02 15:04:05"),
			"answers":    w.Answers,
		}
		list = append(list, item)
	}

	return gin.H{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"page_size":    pageSize,
		"list":         list,
		"has_more":     int64(page) < totalPages,
	}, nil
}
