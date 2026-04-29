package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TaskSuggestion represents AI-generated task recommendations
type TaskSuggestion struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"` // low, medium, high, urgent
	EstimatedHours float64  `json:"estimated_hours"`
	DueDateDays    int      `json:"due_date_days"`
	Assignee       string   `json:"assignee_suggestion"`
	Tags           []string `json:"tags"`
	Dependencies   []string `json:"dependencies"`
	Reasoning      string   `json:"reasoning"`
}

// TaskPriority represents AI-calculated task priority
type TaskPriority struct {
	Score       int     `json:"score"` // 0-100
	Level       string  `json:"level"` // low, medium, high, urgent
	Urgency     float64 `json:"urgency"`
	Importance  float64 `json:"importance"`
	DeadlineRisk string  `json:"deadline_risk"` // on_track, at_risk, overdue_likely
}

// TaskAgent provides AI-powered task management
type TaskAgent struct {
	gatewayURL string
	httpClient *http.Client
}

// NewTaskAgent creates a new task management agent
func NewTaskAgent(gatewayURL string) *TaskAgent {
	return &TaskAgent{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateTaskFromNote creates a structured task from a brief note
func (ta *TaskAgent) GenerateTaskFromNote(ctx context.Context, note string, context map[string]string) (*TaskSuggestion, error) {
	prompt := ta.buildTaskGenerationPrompt(note, context)

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ta.parseTaskSuggestion(response)
}

// PrioritizeTask calculates priority score for a task
func (ta *TaskAgent) PrioritizeTask(ctx context.Context, task *Task, teamWorkload map[string]int) (*TaskPriority, error) {
	prompt := ta.buildPriorityPrompt(task, teamWorkload)

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return ta.parseTaskPriority(response)
}

// SuggestTaskBreakdown breaks down a complex task into subtasks
func (ta *TaskAgent) SuggestTaskBreakdown(ctx context.Context, taskTitle, taskDescription string) ([]TaskSuggestion, error) {
	prompt := fmt.Sprintf(`Break down this complex task into 3-7 actionable subtasks:

Task: %s
Description: %s

For each subtask, provide:
- Clear, actionable title
- Brief description
- Estimated hours
- Suggested priority

Return as JSON array of objects with keys: title, description, estimated_hours, priority`,
		taskTitle, taskDescription)

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var subtasks []TaskSuggestion
	if err := json.Unmarshal([]byte(response), &subtasks); err != nil {
		// Return single task if parsing fails
		subtask, err := ta.parseTaskSuggestion(response)
		if err != nil {
			return nil, err
		}
		return []TaskSuggestion{*subtask}, nil
	}

	return subtasks, nil
}

// PredictTaskCompletion estimates task completion time
func (ta *TaskAgent) PredictTaskCompletion(ctx context.Context, task *Task, historicalData []map[string]interface{}) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Predict completion time for this task:

Task: %s
Description: %s
Priority: %s
Current Status: %s

Based on similar tasks, predict:
1. Estimated hours to complete
2. Likely completion date
3. Risk factors that might delay
4. Confidence level

Return as JSON with keys: estimated_hours, completion_date, risk_factors, confidence`,
		task.Title, task.Description, task.Priority, task.Status)

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var prediction map[string]interface{}
	if err := json.Unmarshal([]byte(response), &prediction); err != nil {
		return map[string]interface{}{"prediction": response}, nil
	}

	return prediction, nil
}

// SuggestTaskAssignment recommends who should handle a task
func (ta *TaskAgent) SuggestTaskAssignment(ctx context.Context, task *Task, teamMembers []map[string]interface{}) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Recommend the best team member for this task:

Task: %s
Description: %s
Required Skills: %s
Priority: %s
Deadline: %s

Team members and their current workload are provided.

Consider:
1. Skills match
2. Current workload
3. Past performance on similar tasks
4. Availability before deadline

Return JSON with keys: recommended_assignee, reasoning, alternative_assignees, workload_concerns`,
		task.Title, task.Description, "", task.Priority, task.DueDate.Format("2006-01-02"))

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var suggestion map[string]interface{}
	if err := json.Unmarshal([]byte(response), &suggestion); err != nil {
		return map[string]interface{}{"recommendation": response}, nil
	}

	return suggestion, nil
}

// GetTaskInsights provides AI insights about task patterns
func (ta *TaskAgent) GetTaskInsights(ctx context.Context, tasks []*Task) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze these %d tasks and provide insights:

Tasks summary:
- Total: %d
- By status: count each status
- By priority: count each priority
- Overdue: tasks past due date

Provide insights on:
1. Workload distribution
2. Bottlenecks or patterns
3. Recommended actions
4. Risk areas

Return as JSON with keys: workload_analysis, bottlenecks, recommendations, risks`,
		len(tasks), len(tasks))

	response, err := ta.callGateway(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var insights map[string]interface{}
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		return map[string]interface{}{"insights": response}, nil
	}

	return insights, nil
}

func (ta *TaskAgent) buildTaskGenerationPrompt(note string, context map[string]string) string {
	contextStr := ""
	for k, v := range context {
		contextStr += fmt.Sprintf("- %s: %s\n", k, v)
	}

	return fmt.Sprintf(`You are a project management expert. Convert this brief note into a structured, actionable task.

User Note: "%s"

Additional Context:
%s

Create a task with:
1. Clear, specific title (action-oriented)
2. Detailed description with context
3. Appropriate priority (low, medium, high, urgent)
4. Realistic time estimate in hours
5. Suggested due date (days from now)
6. Relevant tags for categorization
7. Any dependencies to consider

Return ONLY valid JSON:
{
  "title": "<string>",
  "description": "<string>",
  "priority": "<low|medium|high|urgent>",
  "estimated_hours": <number>,
  "due_date_days": <number>,
  "assignee_suggestion": "<string or empty>",
  "tags": ["<string>"],
  "dependencies": ["<string>"],
  "reasoning": "<string>"
}`, note, contextStr)
}

func (ta *TaskAgent) buildPriorityPrompt(task *Task, teamWorkload map[string]int) string {
	workloadStr := ""
	for name, count := range teamWorkload {
		workloadStr += fmt.Sprintf("- %s: %d tasks\n", name, count)
	}

	return fmt.Sprintf(`Calculate priority score for this task:

Task: %s
Description: %s
Current Priority: %s
Due Date: %s
Status: %s

Team Workload:
%s

Consider:
1. Urgency (days until deadline)
2. Importance (business impact)
3. Dependencies (blocking other work)
4. Effort required
5. Team capacity

Return JSON:
{
  "score": <0-100>,
  "level": "<low|medium|high|urgent>",
  "urgency": <0.0-1.0>,
  "importance": <0.0-1.0>,
  "deadline_risk": "<on_track|at_risk|overdue_likely>"
}`,
		task.Title, task.Description, task.Priority,
		task.DueDate.Format("2006-01-02"), task.Status, workloadStr)
}

func (ta *TaskAgent) parseTaskSuggestion(response string) (*TaskSuggestion, error) {
	var suggestion TaskSuggestion

	if err := json.Unmarshal([]byte(response), &suggestion); err != nil {
		// Extract key info from text response
		suggestion.Title = "New Task"
		suggestion.Description = response
		suggestion.Priority = "medium"
		suggestion.EstimatedHours = 4
		suggestion.DueDateDays = 7
		suggestion.Reasoning = "Parsed from text response"
	}

	// Validate priority
	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "urgent": true}
	if !validPriorities[suggestion.Priority] {
		suggestion.Priority = "medium"
	}

	return &suggestion, nil
}

func (ta *TaskAgent) parseTaskPriority(response string) (*TaskPriority, error) {
	var priority TaskPriority

	if err := json.Unmarshal([]byte(response), &priority); err != nil {
		// Default values
		priority.Score = 50
		priority.Level = "medium"
		priority.Urgency = 0.5
		priority.Importance = 0.5
		priority.DeadlineRisk = "on_track"
	}

	// Validate score
	if priority.Score < 0 {
		priority.Score = 0
	}
	if priority.Score > 100 {
		priority.Score = 100
	}

	// Set level based on score if not provided
	if priority.Level == "" {
		if priority.Score >= 80 {
			priority.Level = "urgent"
		} else if priority.Score >= 60 {
			priority.Level = "high"
		} else if priority.Score >= 40 {
			priority.Level = "medium"
		} else {
			priority.Level = "low"
		}
	}

	return &priority, nil
}

func (ta *TaskAgent) callGateway(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a project management expert. Return responses in valid JSON format."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  800,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/ai", ta.gatewayURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "papabase-service-key")

	resp, err := ta.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(body))
	}

	var gatewayResp map[string]interface{}
	if err := json.Unmarshal(body, &gatewayResp); err != nil {
		return "", err
	}

	if choices, ok := gatewayResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return strings.TrimSpace(content), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no content in gateway response")
}
