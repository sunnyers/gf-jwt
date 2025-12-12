package google_sheets

import (
	"context"
	"fmt"
	"time"

	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/model"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/repository"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	config     config.GoogleSheetsConfig
	service    *sheets.Service
	repository *repository.Repository
}

// NewClient 创建新的Google Sheets客户端
func NewClient(ctx context.Context, cfg config.GoogleSheetsConfig) (*Client, error) {
	// 使用服务账号凭证初始化
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("初始化Google Sheets服务失败: %w", err)
	}

	return &Client{
		config: cfg,
		service: service,
	}, nil
}

// SetRepository 设置仓库实例
func (c *Client) SetRepository(repo *repository.Repository) {
	c.repository = repo
}

// SyncMessages 同步消息到Google Sheets
func (c *Client) SyncMessages(ctx context.Context, messages []model.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 准备数据
	values := make([][]interface{}, 0, len(messages))
	for _, msg := range messages {
		row := []interface{}{
			msg.Timestamp.Format("2006-01-02 15:04:05"),
			msg.ConversationID,
			msg.Platform,
			msg.SenderName,
			msg.SenderType,
			msg.Content,
			msg.ContentType,
			msg.MediaURL,
		}
		values = append(values, row)
	}

	// 写入数据到Google Sheets
	range_ := fmt.Sprintf("%s!A:H", c.config.SheetName)
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.config.SpreadsheetID,
		range_,
		valueRange,
	).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		return fmt.Errorf("写入Google Sheets失败: %w", err)
	}

	return nil
}

// EnsureSheetExists 确保工作表存在
func (c *Client) EnsureSheetExists(ctx context.Context) error {
	// 获取电子表格信息
	spreadsheet, err := c.service.Spreadsheets.Get(c.config.SpreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("获取电子表格信息失败: %w", err)
	}

	// 检查工作表是否存在
	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == c.config.SheetName {
			sheetExists = true
			break
		}
	}

	if !sheetExists {
		// 创建新工作表
		requests := []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: c.config.SheetName,
					},
				},
			},
		}

		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: requests,
		}

		_, err := c.service.Spreadsheets.BatchUpdate(c.config.SpreadsheetID, batchUpdateRequest).Do()
		if err != nil {
			return fmt.Errorf("创建工作表失败: %w", err)
		}

		// 添加表头
		headers := [][]interface{}{
			{"时间", "对话ID", "平台", "发送者", "发送者类型", "内容", "内容类型", "媒体URL"},
		}
		headerRange := fmt.Sprintf("%s!A1:H1", c.config.SheetName)
		valueRange := &sheets.ValueRange{
			Values: headers,
		}

		_, err = c.service.Spreadsheets.Values.Update(
			c.config.SpreadsheetID,
			headerRange,
			valueRange,
		).ValueInputOption("RAW").Do()

		if err != nil {
			return fmt.Errorf("添加表头失败: %w", err)
		}
	}

	return nil
}

// StartSyncTask 启动同步任务
func (c *Client) StartSyncTask(ctx context.Context, repo *repository.Repository) {
	c.repository = repo
	ticker := time.NewTicker(time.Duration(c.config.SyncInterval) * time.Second)
	defer ticker.Stop()

	// 确保工作表存在
	if err := c.EnsureSheetExists(ctx); err != nil {
		fmt.Printf("确保工作表存在失败: %v\n", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 获取未同步的消息
			messages, err := repo.GetUnsyncedMessages(ctx, 100)
			if err != nil {
				fmt.Printf("获取未同步消息失败: %v\n", err)
				continue
			}

			if len(messages) > 0 {
				// 同步消息
				if err := c.SyncMessages(ctx, messages); err != nil {
					fmt.Printf("同步消息失败: %v\n", err)
					continue
				}

				// 标记为已同步
				messageIDs := make([]uint, len(messages))
				for i, msg := range messages {
					messageIDs[i] = msg.ID
				}
				if err := repo.MarkMessagesAsSynced(ctx, messageIDs); err != nil {
					fmt.Printf("标记消息为已同步失败: %v\n", err)
				}
			}
		}
	}
}
