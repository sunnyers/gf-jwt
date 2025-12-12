package web

import (
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/config"
	"github.com/fintech-canada/redbook-wechat-monitoring/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterRoutes 注册路由
func RegisterRoutes(s *ghttp.Server, svc *service.Service, cfg *config.Config) {
	// API路由组
	apiGroup := s.Group("/api")
	{
		// Webhook路由（不需要认证）
		apiGroup.POST("/webhook/redbook", handleRedbookWebhook(svc))
		apiGroup.POST("/webhook/wechat", handleWeChatWebhook(svc))
		apiGroup.GET("/webhook/wechat", handleWeChatWebhookVerify(svc))
		
		// 公开的健康检查
		apiGroup.GET("/health", func(r *ghttp.Request) {
			r.Response.WriteJson(g.Map{"status": "ok"})
		})

		// 需要认证的管理API
		adminGroup := apiGroup.Group("/admin")
		adminGroup.Middleware(authMiddleware(cfg))
		{
			adminGroup.GET("/conversations", handleGetConversations(svc))
			adminGroup.GET("/conversations/:id/messages", handleGetMessages(svc))
			adminGroup.GET("/stats", handleGetStats(svc))
			adminGroup.POST("/sync", handleManualSync(svc))
		}
	}

	// 管理界面路由
	s.Group("/admin", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware(cfg))
		group.ALL("/*", func(r *ghttp.Request) {
			// 返回管理界面HTML
			r.Response.WriteTpl("admin/index.html")
		})
	})

}

// handleRedbookWebhook 处理Redbook Webhook
func handleRedbookWebhook(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// 验证签名
		signature := r.Header.Get("X-Signature")
		timestamp := r.Header.Get("X-Timestamp")
		nonce := r.Header.Get("X-Nonce")

		bodyBytes := r.GetBody()
		if len(bodyBytes) == 0 {
			r.Response.WriteStatus(400, "Invalid body")
			return
		}
		
		if !svc.VerifyRedbookWebhook(signature, timestamp, nonce, bodyBytes) {
			r.Response.WriteStatus(401, "Invalid signature")
			return
		}

		// 处理消息
		if err := svc.ProcessRedbookWebhook(r.GetCtx(), bodyBytes); err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{"status": "ok"})
	}
}

// handleWeChatWebhook 处理WeChat Webhook
func handleWeChatWebhook(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		bodyBytes := r.GetBody()
		if len(bodyBytes) == 0 {
			r.Response.WriteStatus(400, "Invalid body")
			return
		}
		
		if err := svc.ProcessWeChatWebhook(r.GetCtx(), bodyBytes); err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{"status": "ok"})
	}
}

// handleWeChatWebhookVerify 处理WeChat Webhook验证（GET请求）
func handleWeChatWebhookVerify(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		signature := r.Get("signature").String()
		timestamp := r.Get("timestamp").String()
		nonce := r.Get("nonce").String()
		echostr := r.Get("echostr").String()

		result, valid := svc.VerifyWeChatWebhook(signature, timestamp, nonce, echostr)
		if !valid {
			r.Response.WriteStatus(401, "Invalid signature")
			return
		}

		r.Response.Write(result)
	}
}

// handleGetConversations 获取对话列表
func handleGetConversations(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		filters := make(map[string]interface{})
		if platform := r.Get("platform").String(); platform != "" {
			filters["platform"] = platform
		}
		if accountID := r.Get("account_id").String(); accountID != "" {
			filters["account_id"] = accountID
		}
		if status := r.Get("status").String(); status != "" {
			filters["status"] = status
		}

		conversations, err := svc.GetConversations(r.GetCtx(), filters)
		if err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{
			"code": 0,
			"data": conversations,
		})
	}
}

// handleGetMessages 获取消息列表
func handleGetMessages(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		conversationID := r.Get("id").Uint()
		if conversationID == 0 {
			r.Response.WriteStatus(400, "Invalid conversation ID")
			return
		}

		messages, err := svc.GetMessages(r.GetCtx(), conversationID)
		if err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{
			"code": 0,
			"data": messages,
		})
	}
}

// handleGetStats 获取统计信息
func handleGetStats(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		stats, err := svc.GetStats(r.GetCtx())
		if err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{
			"code": 0,
			"data": stats,
		})
	}
}

// handleManualSync 手动触发同步
func handleManualSync(svc *service.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if err := svc.TriggerSync(r.GetCtx()); err != nil {
			r.Response.WriteStatus(500, err.Error())
			return
		}

		r.Response.WriteJson(g.Map{
			"code": 0,
			"message": "同步已触发",
		})
	}
}

// authMiddleware 认证中间件
func authMiddleware(cfg *config.Config) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// 简单的JWT或Session认证
		// 这里简化处理，实际应该使用JWT
		token := r.Header.Get("Authorization")
		if token == "" {
			r.Response.WriteStatus(401, "Unauthorized")
			r.ExitAll()
			return
		}

		// TODO: 验证JWT token
		r.Middleware.Next()
	}
}
