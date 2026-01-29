# PR-Agent 提示词系统完整分析

> 基于 Codium-ai/pr-agent 项目的提示词工程实践分析

---

## 📋 目录

1. [核心提示词概览](#核心提示词概览)
2. [代码审查提示词](#代码审查提示词-pr_reviewer_prompts)
3. [PR 描述生成提示词](#pr-描述生成提示词-pr_description_prompts)
4. [代码建议提示词](#代码建议提示词-pr_code_suggestions_prompts)
5. [问答提示词](#问答提示词-pr_questions_prompts)
6. [其他工具提示词](#其他工具提示词)
7. [提示词工程最佳实践](#提示词工程最佳实践)
8. [如何应用到自己的项目](#如何应用到自己的项目)

---

## 核心提示词概览

PR-Agent 使用 TOML 格式组织提示词，主要包含以下几类：

| 提示词文件 | 功能 | 输出格式 |
|-----------|------|---------|
| `pr_reviewer_prompts.toml` | 代码审查 | YAML (结构化) |
| `pr_description_prompts.toml` | 生成 PR 描述 | YAML (结构化) |
| `pr_code_suggestions_prompts.toml` | 代码改进建议 | YAML (结构化) |
| `pr_questions_prompts.toml` | 回答 PR 相关问题 | 自然语言 |
| `pr_help_prompts.toml` | 帮助文档生成 | Markdown |
| `pr_update_changelog_prompts.toml` | 更新 Changelog | Markdown |

### 核心设计理念

1. **结构化输出** - 使用 Pydantic 模型定义输出格式
2. **Jinja2 模板** - 动态生成提示词
3. **多层次指令** - System Prompt + User Prompt
4. **上下文丰富** - 包含 PR diff、描述、Ticket 信息等

---

## 代码审查提示词 (pr_reviewer_prompts)

### System Prompt 核心结构

````markdown
You are PR-Reviewer, a language model designed to review a Git Pull Request (PR).
Your task is to provide constructive and concise feedback for the PR.
The review should focus on new code added in the PR code diff (lines starting with '+')
````

### 关键组成部分

#### 1️⃣ **Diff 格式说明**

````
## File: 'src/file1.py'
### AI-generated changes summary:
* ...

@@ ... @@ def func1():
__new hunk__
11  unchanged code line0
12  unchanged code line1
13 +new code line2 added
14  unchanged code line3
__old hunk__
 unchanged code line0
 unchanged code line1
-old code line2 removed
 unchanged code line3
````

**设计亮点**：
- ✅ 分离 `__new hunk__` 和 `__old hunk__`
- ✅ 添加行号方便引用
- ✅ 符号前缀 (`+`, `-`, ` `) 明确标识变更类型

#### 2️⃣ **输出数据结构**

使用 Pydantic 定义严格的输出格式：

````python
class KeyIssuesComponentLink(BaseModel):
    relevant_file: str = Field(description="The full file path of the relevant file")
    issue_header: str = Field(description="One or two word title for the issue. For example: 'Possible Bug', etc.")
    issue_content: str = Field(description="A short and concise summary of what should be further inspected and validated during the PR review process for this issue. Do not mention line numbers in this field.")
    start_line: int = Field(description="The start line that corresponds to this issue in the relevant file")
    end_line: int = Field(description="The end line that corresponds to this issue in the relevant file")

class Review(BaseModel):
    key_issues_to_review: List[KeyIssuesComponentLink] = Field("A short and diverse list (0-{{ num_max_findings }} issues) of high-priority bugs, problems or performance concerns introduced in the PR code")
    security_concerns: str = Field(description="Does this PR code introduce vulnerabilities...")
    relevant_tests: str = Field(description="yes/no question: does this PR have relevant tests added or updated ?")
    estimated_effort_to_review_[1-5]: int = Field(description="Estimate, on a scale of 1-5...")
    score: str = Field(description="Rate this PR on a scale of 0-100...")
````

#### 3️⃣ **可选功能模块**

通过 Jinja2 条件语句控制：

````jinja2
{%- if require_security_review %}
    security_concerns: str = Field(...)
{%- endif %}

{%- if require_todo_scan %}
    todo_sections: Union[List[TodoSection], str] = Field(...)
{%- endif %}

{%- if require_can_be_split_review %}
    can_be_split: List[SubPR] = Field(...)
{%- endif %}

{%- if related_tickets %}
    ticket_compliance_check: List[TicketCompliance] = Field(...)
{%- endif %}
````

#### 4️⃣ **User Prompt 结构**

````jinja2
{%- if related_tickets %}
--PR Ticket Info--
{%- for ticket in related_tickets %}
=====
Ticket URL: '{{ ticket.ticket_url }}'
Ticket Title: '{{ ticket.title }}'
Ticket Description:
#####
{{ ticket.body }}
#####
=====
{% endfor %}
{%- endif %}

--PR Info--
Title: '{{title}}'
Branch: '{{branch}}'

PR Description:
======
{{ description|trim }}
======

The PR code diff:
======
{{ diff|trim }}
======

Response (should be a valid YAML, and nothing else):
```yaml
````

### 示例输出

````yaml
review:
  estimated_effort_to_review_[1-5]: |
    3
  score: 89
  relevant_tests: |
    No
  key_issues_to_review:
    - relevant_file: |
        src/utils/validator.py
      issue_header: |
        Possible Bug
      issue_content: |
        The function doesn't handle None values properly, which could lead to AttributeError
      start_line: 45
      end_line: 47
    - relevant_file: |
        src/api/routes.py
      issue_header: |
        Security Concern
      issue_content: |
        SQL query is constructed using string concatenation, vulnerable to SQL injection
      start_line: 89
      end_line: 91
  security_concerns: |
    SQL injection: Line 89-91 in src/api/routes.py uses unsanitized user input directly in SQL query. 
    Recommendation: Use parameterized queries instead.
````

---

## PR 描述生成提示词 (pr_description_prompts)

### System Prompt

````markdown
You are PR-Reviewer, a language model designed to review a Git Pull Request (PR).
Your task is to provide a full description for the PR content: type, description, title, and files walkthrough.
- Focus on the new PR code (lines starting with '+' in the 'PR Git Diff' section).
- Keep in mind that the 'Previous title', 'Previous description' and 'Commit messages' sections may be partial, simplistic, non-informative or out of date.
- The generated title and description should prioritize the most significant changes.
````

### 输出数据结构

````python
class PRType(str, Enum):
    bug_fix = "Bug fix"
    tests = "Tests"
    enhancement = "Enhancement"
    documentation = "Documentation"
    other = "Other"

class FileDescription(BaseModel):
    filename: str = Field(description="The full file path of the relevant file")
    changes_summary: str = Field(description="concise summary of the changes in the relevant file, in bullet points (1-4 bullet points).")
    changes_title: str = Field(description="one-line summary (5-10 words) capturing the main theme of changes in the file")
    label: str = Field(description="a single semantic label that represents a type of code changes that occurred in the File. Possible values (partial list): 'bug fix', 'tests', 'enhancement', 'documentation', 'error handling', 'configuration changes', 'dependencies', 'formatting', 'miscellaneous', ...")

class PRDescription(BaseModel):
    type: List[PRType] = Field(description="one or more types that describe the PR content")
    description: str = Field(description="summarize the PR changes with 1-4 bullet points, each up to 8 words. For large PRs, add sub-bullets for each bullet if needed.")
    title: str = Field(description="a concise and descriptive title that captures the PR's main theme")
    changes_diagram: str = Field(description='a horizontal diagram that represents the main PR changes, in the format of a valid mermaid LR flowchart')
    pr_files: List[FileDescription] = Field(max_items=20, description="a list of all the files that were changed in the PR")
````

### 示例输出

````yaml
type:
- Bug fix
- Enhancement
description: |
  - Fix null pointer exception in user authentication
  - Add input validation for email format
  - Optimize database query performance
  - Update error messages for clarity
title: |
  Fix authentication bugs and improve validation
changes_diagram: |
  ```mermaid
  flowchart LR
    A["User Input"] --> B["Email Validation"]
    B --> C["Authentication Service"]
    C --> D["Database Query"]
    D --> E["Response Handler"]
  ```
pr_files:
- filename: |
    src/auth/validator.py
  changes_summary: |
    - Added email format validation using regex
    - Fixed null pointer exception in validate_user()
  changes_title: |
    Improve input validation and error handling
  label: |
    enhancement
- filename: |
    src/db/queries.py
  changes_summary: |
    - Optimized user lookup query with index
  changes_title: |
    Database query optimization
  label: |
    performance
````

---

## 代码建议提示词 (pr_code_suggestions_prompts)

### System Prompt 核心

````markdown
You are PR-Reviewer, an AI specializing in Pull Request (PR) code analysis and suggestions.
Your task is to examine the provided code diff, focusing on new code (lines prefixed with '+'), 
and offer concise, actionable suggestions to fix possible bugs and problems, and enhance code quality and performance.
````

### 关键指导原则

````markdown
Specific guidelines for generating code suggestions:
- Provide up to {{ num_code_suggestions }} distinct and insightful code suggestions.
- DO NOT suggest implementing changes that are already present in the '+' lines compared to the '-' lines.
- Focus your suggestions ONLY on new code introduced in the PR ('+' lines in '__new hunk__' sections).
- Prioritize suggestions that address potential issues, critical problems, and bugs in the PR code.
- Don't suggest to add docstring, type hints, or comments, to remove unused imports, or to use more specific exception types.
- Be aware that your input consists only of partial code segments (PR diff code), not the complete codebase.
- When mentioning code elements (variables, names, or files) in your response, surround them with backticks (`).
````

### 输出数据结构

````python
class CodeSuggestion(BaseModel):
    relevant_file: str = Field(description="Full path of the relevant file")
    language: str = Field(description="Programming language used by the relevant file")
    existing_code: str = Field(description="A short code snippet, from a '__new hunk__' section after the PR changes, that the suggestion aims to enhance or fix.")
    suggestion_content: str = Field(description="An actionable suggestion to enhance, improve or fix the new code introduced in the PR.")
    improved_code: str = Field(description="A refined code snippet that replaces the 'existing_code' snippet after implementing the suggestion.")
    one_sentence_summary: str = Field(description="A concise, single-sentence overview (up to 6 words) of the suggested improvement.")
    label: str = Field(description="A single, descriptive label that best characterizes the suggestion type. Possible labels include 'security', 'possible bug', 'possible issue', 'performance', 'enhancement', 'best practice', 'maintainability', 'typo'.")

class PRCodeSuggestions(BaseModel):
    code_suggestions: List[CodeSuggestion]
````

### 示例输出

````yaml
code_suggestions:
- relevant_file: |
    src/api/routes.py
  language: |
    python
  existing_code: |
    @app.route('/users/<user_id>')
    def get_user(user_id):
        query = f"SELECT * FROM users WHERE id = {user_id}"
        result = db.execute(query)
  suggestion_content: |
    Use parameterized queries to prevent SQL injection vulnerabilities. Never concatenate user input directly into SQL queries.
  improved_code: |
    @app.route('/users/<user_id>')
    def get_user(user_id):
        query = "SELECT * FROM users WHERE id = ?"
        result = db.execute(query, (user_id,))
  one_sentence_summary: |
    Prevent SQL injection vulnerability
  label: |
    security

- relevant_file: |
    src/utils/helper.py
  language: |
    python
  existing_code: |
    def process_items(items):
        result = []
        for item in items:
            if item['status'] == 'active':
                result.append(item)
        return result
  suggestion_content: |
    Use list comprehension for better performance and readability when filtering lists.
  improved_code: |
    def process_items(items):
        return [item for item in items if item['status'] == 'active']
  one_sentence_summary: |
    Use list comprehension for filtering
  label: |
    performance

- relevant_file: |
    src/models/user.py
  language: |
    python
  existing_code: |
    def validate_email(email):
        if '@' in email:
            return True
        return False
  suggestion_content: |
    Email validation is too simplistic. Use a proper regex pattern or a validation library to ensure email format correctness.
  improved_code: |
    import re
    
    def validate_email(email):
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        return bool(re.match(pattern, email))
  one_sentence_summary: |
    Improve email validation logic
  label: |
    possible bug
````

---

## 问答提示词 (pr_questions_prompts)

### System Prompt

````markdown
You are PR-Reviewer, a language model designed to answer questions about a Git Pull Request (PR).

Your goal is to answer questions\tasks about the new code introduced in the PR (lines starting with '+' in the 'PR Git Diff' section), and provide feedback.
Be informative, constructive, and give examples. Try to be as specific as possible.
Don't avoid answering the questions. You must answer the questions, as best as you can, without adding any unrelated content.
````

### User Prompt 结构

````jinja2
PR Info:

Title: '{{title}}'
Branch: '{{branch}}'

{%- if description %}
Description:
======
{{ description|trim }}
======
{%- endif %}

The PR Git Diff:
======
{{ diff|trim }}
======

The PR Questions:
======
{{ questions|trim }}
======

Response to the PR Questions:
````

### 使用场景

- 询问特定代码片段的作用
- 解释复杂的逻辑
- 询问为什么做某个改动
- 技术决策的原因

---

## 其他工具提示词

### 1. Help 文档生成 (pr_help_prompts.toml)

生成帮助文档，解释 PR 中的关键概念。

### 2. Changelog 更新 (pr_update_changelog_prompts.toml)

自动生成 CHANGELOG.md 条目。

### 3. 自定义标签 (pr_custom_labels.toml)

为 PR 自动打标签（bug、feature、breaking-change 等）。

### 4. 文档生成 (pr_add_docs.toml)

为代码自动添加文档注释。

---

## 提示词工程最佳实践

### 1️⃣ **结构化输出的设计**

**✅ 优点**：
- 可解析、可验证
- 类型安全
- 易于集成到 CI/CD

**实现方式**：

````python
# 1. 使用 Pydantic 定义 Schema
class CodeSuggestion(BaseModel):
    relevant_file: str
    existing_code: str
    improved_code: str
    label: str

# 2. 在 Prompt 中明确声明
"""
The output must be a YAML object equivalent to type $CodeSuggestion, according to the following Pydantic definitions:
=====
class CodeSuggestion(BaseModel):
    ...
=====
"""

# 3. 提供示例输出
"""
Example output:
```yaml
code_suggestions:
- relevant_file: |
    src/file.py
  existing_code: |
    ...
```
"""
````

### 2️⃣ **上下文分层设计**

````
Level 1: System Prompt (角色定义)
  ↓
Level 2: Task Description (任务描述)
  ↓
Level 3: Input Format (输入格式说明)
  ↓
Level 4: Output Schema (输出格式定义)
  ↓
Level 5: Guidelines (具体指导原则)
  ↓
Level 6: Examples (示例)
  ↓
Level 7: User Input (实际数据)
````

### 3️⃣ **使用 Jinja2 模板实现灵活性**

````jinja2
{%- if extra_instructions %}
Extra instructions from the user:
======
{{ extra_instructions }}
======
{% endif %}

{%- if require_security_review %}
    security_concerns: str = Field(...)
{%- endif %}

{%- for ticket in related_tickets %}
Ticket Title: '{{ ticket.title }}'
{%- endfor %}
````

### 4️⃣ **明确的约束和边界**

````markdown
Constraints:
- DO NOT suggest implementing changes that are already present
- Focus ONLY on new code ('+' lines)
- Avoid suggestions that might duplicate existing functionality
- When quoting code, use backticks (`) instead of single quote (')
- Provide up to {{ num_code_suggestions }} suggestions
````

### 5️⃣ **Few-shot Learning**

````yaml
Example output:
```yaml
code_suggestions:
- relevant_file: |
    src/file1.py
  language: |
    python
  existing_code: |
    ...
  suggestion_content: |
    ...
  improved_code: |
    ...
  one_sentence_summary: |
    ...
  label: |
    ...
```
````

---

## 如何应用到自己的项目

### 步骤 1：创建提示词模板文件

````typescript
// src/prompts/codeReviewPrompts.ts

export const CODE_REVIEW_SYSTEM_PROMPT = `
You are a code review expert specializing in TypeScript and React.
Your task is to review code changes and provide constructive feedback.

Focus on:
1. Potential bugs and logic errors
2. Security vulnerabilities
3. Performance issues
4. Code maintainability
5. Best practices violations

Output Format: JSON
{
  "summary": "Overall assessment",
  "issues": [
    {
      "severity": "critical|high|medium|low",
      "type": "bug|security|performance|style",
      "file": "file path",
      "line": number,
      "message": "Issue description",
      "suggestion": "How to fix"
    }
  ],
  "score": 0-100
}
`;

export const CODE_REVIEW_USER_PROMPT = (params: {
  fileName: string;
  diff: string;
  description?: string;
}) => `
File: ${params.fileName}

${params.description ? `Description: ${params.description}\n` : ''}

Code Diff:
\`\`\`diff
${params.diff}
\`\`\`

Please review the code changes above.
`;
````

### 步骤 2：集成到 Agent

````typescript
import Anthropic from '@anthropic-ai/sdk';
import { CODE_REVIEW_SYSTEM_PROMPT, CODE_REVIEW_USER_PROMPT } from './prompts/codeReviewPrompts';

export class CodeReviewAgent {
  private client: Anthropic;

  constructor(apiKey: string) {
    this.client = new Anthropic({ apiKey });
  }

  async reviewCode(params: {
    fileName: string;
    diff: string;
    description?: string;
  }): Promise<CodeReviewResult> {
    const response = await this.client.messages.create({
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 4096,
      system: CODE_REVIEW_SYSTEM_PROMPT,
      messages: [
        {
          role: 'user',
          content: CODE_REVIEW_USER_PROMPT(params),
        },
      ],
    });

    const content = response.content[0];
    if (content.type === 'text') {
      return JSON.parse(content.text);
    }

    throw new Error('Unexpected response format');
  }
}

interface CodeReviewResult {
  summary: string;
  issues: Array<{
    severity: 'critical' | 'high' | 'medium' | 'low';
    type: 'bug' | 'security' | 'performance' | 'style';
    file: string;
    line: number;
    message: string;
    suggestion: string;
  }>;
  score: number;
}
````

### 步骤 3：创建配置文件

````toml
# config/review.toml

[review]
max_issues = 10
min_severity = "medium"
focus_areas = ["security", "performance", "bugs"]

[prompts]
enable_security_scan = true
enable_performance_check = true
require_tests = true

[output]
format = "json"
include_line_numbers = true
add_severity_emoji = true
````

### 步骤 4：实现模板系统

````typescript
import Handlebars from 'handlebars';
import fs from 'fs';

export class PromptTemplate {
  private template: HandlebarsTemplateDelegate;

  constructor(templatePath: string) {
    const templateContent = fs.readFileSync(templatePath, 'utf-8');
    this.template = Handlebars.compile(templateContent);
  }

  render(data: Record<string, any>): string {
    return this.template(data);
  }
}

// 使用示例
const reviewPromptTemplate = new PromptTemplate('./prompts/review.hbs');

const prompt = reviewPromptTemplate.render({
  fileName: 'src/utils.ts',
  diff: '...',
  enableSecurity: true,
  maxIssues: 10,
});
````

---

## 总结

### PR-Agent 提示词系统的核心优势

1. ✅ **结构化输出** - 使用 Pydantic 强制类型约束
2. ✅ **模块化设计** - 通过 Jinja2 实现可配置的提示词
3. ✅ **领域特化** - 针对不同任务使用专门的提示词
4. ✅ **示例驱动** - Few-shot learning 提升输出质量
5. ✅ **约束明确** - 清晰定义能做什么、不能做什么
6. ✅ **上下文丰富** - 提供充足的背景信息（Ticket、Commit、Diff）

### 可以学习借鉴的点

1. **使用 TOML 管理提示词** - 比硬编码在代码中更易维护
2. **Pydantic Schema 定义** - 确保输出格式一致性
3. **分层提示词设计** - System Prompt + User Prompt + Examples
4. **Jinja2 模板语法** - 实现动态提示词生成
5. **明确的约束条件** - 减少模型的发散性
6. **上下文注入策略** - 提供恰当的背景信息

---

## 附录：完整提示词模板示例

见下一页...
