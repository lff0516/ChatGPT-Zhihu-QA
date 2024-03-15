package v1

import (
	"context"
	"fmt"
	"log"
	"os"
	"qa/model"
	"qa/serializer"
	v1 "qa/service/v1/answer"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type QesAddService struct {
	Title   string `form:"title" json:"title" binding:"required"`
	Content string `form:"content" json:"content"`
}

func (qesAddService *QesAddService) QuestionAdd(user *model.User) *serializer.Response {
	qes := model.Question{
		UserID:  user.ID,
		Title:   qesAddService.Title,
		Content: qesAddService.Content,
	}

	if err := model.DB.Create(&qes).Error; err != nil {
		return serializer.ErrorResponse(serializer.CodeDatabaseError)
	}

	// 成功创建问题后，提取并输出 question.ID
	createdQuestion := serializer.BuildQuestionResponse(&qes, user.ID)
	if createdQuestion != nil && createdQuestion.Question != nil {
		questionID := createdQuestion.Question.ID
		fmt.Println("Newly created question ID:", questionID)

		// 将 questionID 和 qesAddService.Content 传入 AddllmAnswer 函数
		AddllmAnswer(questionID, qesAddService.Content, user)
	}

	return serializer.OkResponse(createdQuestion)
}

// 添加llm回答，传入的参数有问题的内容，问题的ID，在这个方法中进行RAG，通过问题内容输入到图数据库中检索相关的上下文，然后将问题与相关的上下文
// 一起通过langchain-go框架输入到对话大模型中，输出的内容作为回答，通过service.AddAnswer放入到mysql数据库中，等待客户端的抽取，

func AddllmAnswer(questionID uint, contentValue string, user *model.User) {

	// 设置 OpenAI API 访问凭证和基本地址
	os.Setenv("OPENAI_API_KEY", "sk-zk236783dee1d240ba89d8402f0e44c22f065b5f21ea7300")
	os.Setenv("OPENAI_API_BASE", "https://flag.smarttrot.com/v1")

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	rules := `再用至少100个字延展下对上面问题的回答。需要标注其是由AI生成的`
	prompt := contentValue + rules
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Newly created prompt:", prompt, "Newly created completion:", completion)

	// 解析参数
	var service v1.AddAnswerService
	service.Content = completion // llm生成的回答

	// 执行service.AddAnswer方法
	res := service.AddAnswer(user, questionID)
	fmt.Println("service AddAnswer:", res)
}
