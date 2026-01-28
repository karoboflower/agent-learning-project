// Agent Types
export interface AgentConfig {
  apiKey: string;
  model: string;
  temperature?: number;
  maxTokens?: number;
}

export interface AgentResponse {
  content: string;
  usage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
}

// Code Review Types
export interface CodeReviewRequest {
  code: string;
  language: string;
  context?: string;
}

export interface CodeReviewResponse {
  issues: CodeIssue[];
  suggestions: string[];
  overallQuality: number;
}

export interface CodeIssue {
  line: number;
  severity: 'error' | 'warning' | 'info';
  message: string;
  suggestion?: string;
}

// Refactor Types
export interface RefactorRequest {
  code: string;
  language: string;
  goal: string;
}

export interface RefactorResponse {
  refactoredCode: string;
  changes: RefactorChange[];
  explanation: string;
}

export interface RefactorChange {
  type: 'add' | 'remove' | 'modify';
  description: string;
  line?: number;
}

// Tech Stack Types
export interface TechStackRequest {
  projectDescription: string;
  requirements: string[];
  constraints?: string[];
}

export interface TechStackResponse {
  recommendations: TechStackRecommendation[];
  reasoning: string;
}

export interface TechStackRecommendation {
  category: string;
  technology: string;
  reason: string;
  alternatives?: string[];
}
