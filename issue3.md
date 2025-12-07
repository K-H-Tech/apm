
in internal/repository/prd_repository.go
1- Handle UUID parse error instead of discarding it.

If template_id contains an invalid UUID string in the database, this silently assigns uuid.Nil to TemplateID, potentially masking data corruption.

     if templateID.Valid {
-		tid, _ := uuid.Parse(templateID.String)
+		tid, err := uuid.Parse(templateID.String)
+		if err != nil {
+			return nil, fmt.Errorf("invalid template_id UUID: %w", err)
+		}
         prd.TemplateID = &tid
     }
Note: This pattern is repeated in ListByOrganization (line 232), ListByOwner (line 296), ListByStatus (line 360), and Search (line 545).


2- Escape LIKE/ILIKE special characters in search pattern.

The % and _ characters have special meaning in LIKE patterns. If a user searches for "100%", it becomes %100%% which matches unintended strings. Escape special characters before wrapping with wildcards.

+// escapeLikePattern escapes LIKE special characters
+func escapeLikePattern(s string) string {
+	s = strings.ReplaceAll(s, "\\", "\\\\")
+	s = strings.ReplaceAll(s, "%", "\\%")
+	s = strings.ReplaceAll(s, "_", "\\_")
+	return s
+}
+
 // In Search method:
-searchPattern := "%" + query + "%"
+searchPattern := "%" + escapeLikePattern(query) + "%"
Note: You'll need to add "strings" to the imports.

pkg/atlassian/oauth.go
1- Interface implementation inconsistency.

The StateGenerator interface defines both Generate() and Validate() methods, but SimpleStateStore only implements Validate() (it has Store() instead of Generate()). This means SimpleStateStore does not implement the StateGenerator interface, which could lead to type assertion failures or compilation errors when trying to use it as a StateGenerator.

Consider either:

Adding a Generate() method to SimpleStateStore that generates a cryptographically secure random string, or
Removing the StateGenerator interface if it's not being used
2- Prevent panic on multiple Stop() calls.

The Stop() method closes the stopChan without checking if it's already closed. Calling Stop() multiple times will panic. Use sync.Once to ensure the channel is only closed once.

Apply this diff:

 type SimpleStateStore struct {
     mu       sync.Mutex
     states   map[string]time.Time
     ttl      time.Duration
     stopChan chan struct{}
+	stopOnce sync.Once
 }

 // Stop stops the cleanup goroutine. Should be called when the store is no longer needed.
 func (s *SimpleStateStore) Stop() {
-	close(s.stopChan)
+	s.stopOnce.Do(func() {
+		close(s.stopChan)
+	})
 }

in jwt.go
1- Validate secret key strength for HS512.

For HS512, the secret key should be at least 64 bytes (512 bits) to match the algorithm's security strength. Weak or short secrets are vulnerable to brute-force attacks.

Consider adding validation:

 func NewJWT(secretKey string) *JWT {
+	if len(secretKey) < 64 {
+		panic("JWT secret key must be at least 64 bytes for HS512")
+	}
     return &JWT{
         secretKey: []byte(secretKey),
     }
 }
Alternatively, validate at the configuration level where the secret is loaded.

2- Fix critical bug in error handling and add token validity check.

There are several issues in the Verify method:

Critical bug at line 51: The error message uses err with %w, but err refers to the parsing error from line 39. If we reach line 51, err is nil (we already returned if it wasn't at line 47). This produces incorrect error messages and tries to wrap a nil error.

Missing validity check: After successful parsing, verify that jwtToken.Valid is true and jwtToken is not nil before extracting claims.

Unused context: The ctx parameter is never used.

Apply this diff to fix the issues:

 func (j *JWT) Verify(ctx context.Context, token string) (*models.JWT, error) {
     jwtToken, err := jwt.ParseWithClaims(token, &models.JWT{}, func(t *jwt.Token) (interface{}, error) {
         if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
             return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
         }
         return j.secretKey, nil
     })
     if err != nil {
         return nil, fmt.Errorf("JWT. error at parsing the token: %w", err)
     }
+
+	if jwtToken == nil || !jwtToken.Valid {
+		return nil, fmt.Errorf("JWT. token is invalid")
+	}

     claims, ok := jwtToken.Claims.(*models.JWT)
     if !ok {
-		return nil, fmt.Errorf("JWT. error at asserting the claims to models.JWT: %w", err)
+		return nil, fmt.Errorf("JWT. error at asserting the claims to models.JWT")
     }

     return claims, nil
 }
in  pkg/llm/prompts/story_writer.go
1- Inconsistency between prompt and schema for Fibonacci values.

The schema enum includes 21 in the Fibonacci sequence, but the UserStoryGeneratorPrompt on line 26 only mentions values up to 13. This inconsistency may cause confusion about which values are acceptable.

Apply this diff to align the schema with the prompt:

 "estimate_points": map[string]interface{}{
     "type":        "integer",
-	"enum":        []int{1, 2, 3, 5, 8, 13, 21},
+	"enum":        []int{1, 2, 3, 5, 8, 13},
     "description": "Fibonacci story points",
 },
Alternatively, if you want to support 21, update line 26 in the prompt:

-4. Suggest story point estimates using Fibonacci (1, 2, 3, 5, 8, 13)
+4. Suggest story point estimates using Fibonacci (1, 2, 3, 5, 8, 13, 21)
2-Fix trailing comma in template range loops.

The range loops for Goals and PainPoints append a comma after each item, including the last one. This produces output like "goal1, goal2, goal3, " with an unwanted trailing comma and space.

Apply this diff to fix the trailing comma issue:

-  Goals: {{range .Goals}}{{.}}, {{end}}
-  Pain Points: {{range .PainPoints}}{{.}}, {{end}}
+  Goals: {{range $i, $goal := .Goals}}{{if $i}}, {{end}}{{$goal}}{{end}}
+  Pain Points: {{range $i, $pain := .PainPoints}}{{if $i}}, {{end}}{{$pain}}{{end}}

in pkg/llm/provider.go
1- Add mutex protection to prevent concurrent map access race condition.

The providers map is accessed and modified without synchronization. If GetProvider is called concurrently from multiple goroutines, this creates a race condition that could lead to:

Multiple instances of the same provider being created
Map corruption and potential panics
Apply this diff to add mutex protection:

 type ProviderFactory struct {
     config    *Config
     providers map[ProviderType]contract.LLMProvider
+	mu        sync.RWMutex
 }
Add import for sync:

 import (
     "context"
     "errors"
+	"sync"
     "time"

     "github.com/K-H-Tech/apm/internal/contract"
 )
Protect the map access:

 func (f *ProviderFactory) GetProvider(providerType ProviderType) (contract.LLMProvider, error) {
     // Check if already created
+	f.mu.RLock()
-	if provider, ok := f.providers[providerType]; ok {
+	provider, ok := f.providers[providerType]
+	f.mu.RUnlock()
+	if ok {
         return provider, nil
     }

+	f.mu.Lock()
+	defer f.mu.Unlock()
+
+	// Double-check after acquiring write lock
+	if provider, ok := f.providers[providerType]; ok {
+		return provider, nil
+	}
+
     // Create provider
     var provider contract.LLMProvider
     var err error

     switch providerType {
     case ProviderOpenAI:
         if f.config.OpenAIAPIKey == "" {
             return nil, ErrProviderNotConfigured
         }
         provider, err = NewOpenAIProvider(f.config)
     case ProviderAnthropic:
         if f.config.AnthropicAPIKey == "" {
             return nil, ErrProviderNotConfigured
         }
         provider, err = NewAnthropicProvider(f.config)
     default:
         return nil, errors.New("unknown provider type: " + string(providerType))
     }

     if err != nil {
         return nil, err
     }

     f.providers[providerType] = provider
     return provider, nil
 }
2- Implement retry logic or remove the unused maxRetries field.

The maxRetries field is stored in the struct but never used. Additionally, the current fallback logic has two concerns:

No retry attempts on the primary provider before falling back
Fallback triggers on all errors, including ones that would fail on the fallback provider too (e.g., ErrContextTooLong, ErrProviderNotConfigured)
Option 1: Remove unused field if retry logic isn't needed

 type MultiProvider struct {
     primary   contract.LLMProvider
     fallback  contract.LLMProvider
-	maxRetries int
 }

-func NewMultiProvider(primary, fallback contract.LLMProvider, maxRetries int) *MultiProvider {
+func NewMultiProvider(primary, fallback contract.LLMProvider) *MultiProvider {
     return &MultiProvider{
         primary:   primary,
         fallback:  fallback,
-		maxRetries: maxRetries,
     }
 }
Option 2: Implement retry logic and smarter fallback

 func (m *MultiProvider) GenerateText(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
-	// Try primary provider
-	resp, err := m.primary.GenerateText(ctx, req)
-	if err == nil {
-		return resp, nil
+	var lastErr error
+
+	// Try primary provider with retries
+	for attempt := 0; attempt <= m.maxRetries; attempt++ {
+		resp, err := m.primary.GenerateText(ctx, req)
+		if err == nil {
+			return resp, nil
+		}
+		lastErr = err
+
+		// Don't retry on non-retryable errors
+		if errors.Is(err, ErrContextTooLong) || errors.Is(err, ErrProviderNotConfigured) {
+			break
+		}
     }

     // If fallback is configured, try it
-	if m.fallback != nil {
+	if m.fallback != nil && !errors.Is(lastErr, ErrContextTooLong) && !errors.Is(lastErr, ErrProviderNotConfigured) {
         return m.fallback.GenerateText(ctx, req)
     }

-	return nil, err
+	return nil, lastErr
 }

in pkg/priority/weighted.go
1- Consider handling tied scores in ranking.

Items with identical scores currently receive different sequential ranks. In many ranking contexts, tied items should share the same rank with subsequent ranks adjusted accordingly (e.g., scores [100, 95, 95, 90] → ranks [1, 2, 2, 4] instead of [1, 2, 3, 4]).

If tie handling is desired, you could modify the logic:

 func RankItems(items []WeightedScoredItem) []RankedItem {
     // Sort by score descending
     sort.Slice(items, func(i, j int) bool {
         return items[i].Score.TotalScore > items[j].Score.TotalScore
     })

     ranked := make([]RankedItem, len(items))
+	currentRank := 1
     for i, item := range items {
+		if i > 0 && items[i].Score.TotalScore < items[i-1].Score.TotalScore {
+			currentRank = i + 1
+		}
         ranked[i] = RankedItem{
-			Rank:  i + 1,
+			Rank:  currentRank,
             ID:    item.ID,
             Name:  item.Name,
             Score: item.Score,
         }
     }

     return ranked
 }
2- Fix: NewWeight should reflect the actual normalized weight used in calculations.

The sensitivity analysis has a logic flaw:

When adjusting weights proportionally (line 474-482), some weights may be clamped to 0, causing the sum to differ from 100
A calculator with normalization is used (line 485) to handle this, which rescales all weights
However, NewWeight (line 493) records the pre-normalization weight, not the actual weight used in the calculation
Example: With criteria [5, 5, 90], increasing the first by +20:

Pre-normalization: [25, 0, 80] (sum = 105)
After normalization: [23.81, 0, 76.19] (sum = 100)
NewWeight reports 25, but actual weight used is 23.81
Actual change is 18.81, not 20
This makes the sensitivity analysis results misleading.

To fix, capture and record the actual normalized weights:

         // Calculate with normalization to handle weight sum variations
         calcWithNorm := NewWeightedCalculatorWithNormalization()
         modifiedResult, err := calcWithNorm.Calculate(&WeightedParams{Criteria: modifiedCriteria})
         if err != nil {
             continue
         }
+
+		// Find the actual normalized weight for the target criterion
+		var actualNewWeight float64
+		for _, r := range modifiedResult.Criteria {
+			if r.Name == criterionName {
+				actualNewWeight = r.Weight
+				break
+			}
+		}

         results = append(results, ScoreVariation{
             WeightChange: change,
-			NewWeight:    newWeight,
+			NewWeight:    actualNewWeight,
             NewScore:     modifiedResult.TotalScore,
             ScoreChange:  modifiedResult.TotalScore - baseResult.TotalScore,
         })
in tools/postgresql_connection.go
1- Critical security issue: sslmode=disable transmits credentials and data in plain text.

The hardcoded sslmode=disable means all traffic between the application and PostgreSQL (including credentials and sensitive data) is unencrypted and vulnerable to network sniffing.

Make sslmode configurable via config.PostgreSQL and default to a secure option:

-	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
+	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
         pgCfg.Host,
         pgCfg.Port,
         pgCfg.Username,
         pgCfg.Password,
         pgCfg.DBName,
+		pgCfg.SSLMode, // Add SSLMode field to config with default "require" or "verify-full"
     )
Additionally, consider adding connect_timeout to prevent indefinite connection hangs:

connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
    pgCfg.Host,
    pgCfg.Port,
    pgCfg.Username,
    pgCfg.Password,
    pgCfg.DBName,
    pgCfg.SSLMode,
    pgCfg.ConnectTimeout, // e.g., 10 seconds
)
2- Inconsistent error handling: panic instead of returning error.

The function signature returns an error, but log.Panicln(err) terminates the program instead of returning the error to the caller. This violates the error-handling contract and prevents graceful error recovery.

Return the error instead of panicking:

     d, err := sql.Open("postgres", connStr)
     if err != nil {
-		log.Panicln(err)
+		return nil, fmt.Errorf("failed to open database: %w", err)
     }
in tools/redis_connection.go
1- Add a timeout to the Ping context.

Using context.Background() without a timeout means the Ping operation could block indefinitely if Redis is unreachable or slow. This can cause application startup or reconnection attempts to hang.

Apply this diff to add a timeout:

+	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
+	defer cancel()
+
-	if err := rdb.Ping(context.Background()).Err(); err != nil {
+	if err := rdb.Ping(ctx).Err(); err != nil {
         return nil, err
     }
You'll also need to import the time package:

 import (
     "context"
+	"time"

     "github.com/K-H-Tech/apm/internal/config"
2- Add missing configuration options to the standard Redis client.

The standard client only configures Addr, Password, and DB, while the sentinel client also configures MaxRetries, PoolSize, DialTimeout, ReadTimeout, WriteTimeout, and PoolTimeout. This inconsistency means timeout and pool settings from redisCfg are ignored when not using sentinel mode, potentially leading to unbounded resource usage or indefinite blocking.

Apply this diff to add the missing configuration options:

     } else {
         rdb = redis.NewClient(&redis.Options{
-			Addr:     redisCfg.Addr,
-			Password: redisCfg.Password,
-			DB:       redisCfg.DB,
+			Addr:         redisCfg.Addr,
+			Password:     redisCfg.Password,
+			DB:           redisCfg.DB,
+			MaxRetries:   redisCfg.MaxRetries,
+			PoolSize:     redisCfg.PoolSize,
+			DialTimeout:  redisCfg.DialTimeout,
+			ReadTimeout:  redisCfg.ReadTimeout,
+			WriteTimeout: redisCfg.WriteTimeout,
+			PoolTimeout:  redisCfg.PoolTimeout,
         })
     }
in internal/models/discovery.go
1- Critical: Inconsistent receiver type with DiscoveryContent.Value().

AIInsights.Value() uses a pointer receiver while DiscoveryContent.Value() (line 63) uses a value receiver. This inconsistency creates confusion and maintenance issues within the same file.

For consistency, apply this diff:

-func (a *AIInsights) Value() (driver.Value, error) {
+func (a AIInsights) Value() (driver.Value, error) {
-	if a == nil {
-		return nil, nil
-	}
-	return json.Marshal(a)
+	return json.Marshal(&a)
 }
Alternatively, if nil handling is required, consider making DiscoveryContent.Value() also use a pointer receiver for consistency. The typical pattern is to use value receivers for Value() and pointer receivers for Scan().

2-Use UUID types for entity references instead of strings.

Multiple fields represent references to other entities but use string type: LinkedSolutions, ParentOpportunity, ChildOpportunities (in OpportunityDetails), and TargetOpportunity (in SolutionDetails). This weakens type safety and referential integrity.

Apply this diff to use proper UUID types:

 type OpportunityDetails struct {
     Description       string   `json:"description"`
     TargetUsers       []string `json:"target_users,omitempty"`
     PotentialImpact   string   `json:"potential_impact,omitempty"`
-	LinkedSolutions   []string `json:"linked_solutions,omitempty"`
-	ParentOpportunity string   `json:"parent_opportunity,omitempty"` // For opportunity trees
-	ChildOpportunities []string `json:"child_opportunities,omitempty"`
+	LinkedSolutions   []uuid.UUID `json:"linked_solutions,omitempty"`
+	ParentOpportunity *uuid.UUID  `json:"parent_opportunity,omitempty"` // For opportunity trees
+	ChildOpportunities []uuid.UUID `json:"child_opportunities,omitempty"`
 }

 type SolutionDetails struct {
     Description     string   `json:"description"`
-	TargetOpportunity string  `json:"target_opportunity,omitempty"`
+	TargetOpportunity *uuid.UUID  `json:"target_opportunity,omitempty"`
     Assumptions     []string `json:"assumptions,omitempty"`
     Risks           []string `json:"risks,omitempty"`
     EffortEstimate  string   `json:"effort_estimate,omitempty"`
     Validated       bool     `json:"validated"`
     ValidationNotes string   `json:"validation_notes,omitempty"`
 }
3- Use UUID array type for LinkedPRDIDs instead of StringArray.

The LinkedPRDIDs field uses pq.StringArray (string slice), but these represent UUID references. This undermines type safety and allows invalid UUIDs to be stored without compile-time checks.

Consider using a custom UUID array type with proper validation:

 type DiscoveryItem struct {
     ID             uuid.UUID       `json:"id" db:"id"`
     OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
     Type           DiscoveryType   `json:"type" db:"type"`
     Title          string          `json:"title" db:"title"`
     Content        DiscoveryContent `json:"content" db:"content"`
     Tags           pq.StringArray  `json:"tags" db:"tags"`
-	LinkedPRDIDs   pq.StringArray  `json:"linked_prd_ids" db:"linked_prd_ids"`
+	LinkedPRDIDs   []uuid.UUID     `json:"linked_prd_ids" db:"linked_prd_ids"`
     AIInsights     *AIInsights     `json:"ai_insights,omitempty" db:"ai_insights"`
     CreatedBy      *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
     CreatedAt      time.Time       `json:"created_at" db:"created_at"`
     UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
 }
You may need to implement a custom scanner/valuer or use pq.Array(&item.LinkedPRDIDs) in your database queries.


in .claude/commands/senior-data-analyst.md
1- PostgreSQL-specific DATE_TRUNC syntax is incompatible with MySQL and BigQuery.

The expertise areas (line 7) claim fluency across PostgreSQL, MySQL, and BigQuery, but this query uses DATE_TRUNC, which doesn't exist in MySQL or work identically in BigQuery. For a persona that spans multiple databases, provide either:

Database-agnostic syntax, or
Clearly marked examples for each database with equivalent implementations.
For PostgreSQL (current query):

DATE_TRUNC('week', created_at)
For MySQL:

DATE_SUB(created_at, INTERVAL DAYOFWEEK(created_at)-2 DAY)
For BigQuery:

DATE_TRUNC(CAST(created_at AS DATE), WEEK)
Consider including a comment flagging the database-specific nature or providing multi-database variants.

2- SQL example doesn't demonstrate the claimed date range best practice.

The comment on line 58 states "Always include date ranges explicitly," but the code block lacks a WHERE clause showing date filtering. Expand the example to demonstrate this best practice, or adjust the comment to match the example intent.

Consider revising the example to something like:

-- Use CTEs for readability
WITH filtered_data AS (
    SELECT *
    FROM source_table
    WHERE created_at >= DATE '2024-01-01'
      AND created_at < DATE '2024-02-01'
),
aggregated AS (
    SELECT ...
    FROM filtered_data
)
SELECT * FROM aggregated;
This demonstrates explicit date range filtering while maintaining the CTE pattern.

in redis.md
1- Don't ignore errors in documentation examples.

Line 429 silences the json.Marshal error, which can hide bugs. Documentation should model error handling best practices. Even if error handling would complicate the example, add a brief comment explaining why it's omitted.

Apply this diff to add error handling:

  // Store in cache
  data, _ := json.Marshal(user)
+ data, err := json.Marshal(user)
+ if err != nil {
+     // Log and continue (cache miss is non-fatal)
+     log.Printf("failed to marshal user: %v", err)
+     return user, nil
+ }
  rdb.Set(ctx, cacheKey, data, time.Hour)
in .claude/rules/viper.md
1- Flag potential goroutine leak risk in remote config watching pattern.

The polling pattern shown here can cause goroutine leaks because WatchRemoteConfig creates goroutines without providing cancel channels, meaning spawned watchers never terminate. This is a known limitation when using Viper's remote watch feature.

For production use, consider:

Adding a context and graceful shutdown mechanism
Monitoring goroutine counts
Implementing maximum retry limits or circuit breakers
Consider noting this caveat in the documentation.

Would you like me to provide a safer remote config watching pattern with proper goroutine lifecycle management?

in the main.go
Unused AI rate limiter middleware group.

The aiRoutes group is created with the AI rate limiter middleware, but the comment indicates AI routes are already registered via prdHandler.RegisterRoutes. This means the stricter AI rate limiting may not be applied to AI endpoints as intended, and aiRateLimitHandler creates a cleanup goroutine that serves no purpose.

Either remove the unused group or ensure AI routes are registered under this group.

-	// AI routes with stricter rate limiting
-	aiRoutes := api.Group("/ai")
-	aiRateLimitHandler, aiRateLimitCleanup := middleware.AIRateLimitMiddleware()
-	aiRoutes.Use(aiRateLimitHandler)
-	// AI routes are already registered via prdHandler.RegisterRoutes
+	// Note: AI routes are registered via prdHandler.RegisterRoutes
+	// If stricter rate limiting is needed, consider registering AI routes here instead
in the dockerfile :
Reduce package list to only runtime essentials.

The installation includes packages that appear to be development/debugging tools rather than production runtime dependencies:

redis: Why does a server binary need Redis installed? If it's needed as a cache/message broker, it should be in a separate container/service.
curl, bind-tools, tcpdump: Network debugging tools.
yq, jq: Data processing utilities.
busybox-extras, zip, unzip: Utilities without clear production need.
This bloats the image significantly (~50–100+ MB of unnecessary packages), expands the attack surface, and violates security best practices. Keep only what the binary needs to run:

Essential: tzdata, bash, ca-certificates, gcompat
Others: Only if they're hard runtime dependencies (e.g., runtime libraries or config tools).
Suggested minimal replacement:

RUN apk update && \
    apk add --update --no-cache \
    tzdata \
    bash \
    ca-certificates \
    gcompat
Clarify which of the other packages are actual production dependencies vs. debugging/development tools.

in oauth.go
Use a concrete type instead of interface{} for expiresAt.

The expiresAt interface{} parameter loses type safety and requires runtime type assertions. This should be a concrete type.

-	UpdateAtlassianTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, expiresAt interface{}) error
+	UpdateAtlassianTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, expiresAt time.Time) error
Note: You'll need to add "time" to the imports.

in middleware.go
Fix typo in comment.

"ad necessary" should be "and necessary".

Apply this diff:

-// Middlewares will implement all required ad necessary middlewares for the http server
+// Middlewares will implement all required and necessary middlewares for the http server

in jira.go
1- Add validation tag for required field.

IssueKey should have binding:"required" tag if this struct is used for API request validation, as bulk updates require identifying which issue to update.

Apply this diff:

 type JiraBulkUpdate struct {
-	IssueKey string          `json:"issue_key"`
+	IssueKey string          `json:"issue_key" binding:"required"`
     Update   JiraIssueUpdate `json:"update"`
 }

in jwt_token.go
Critical: Inverted logic in IsExpired method.

The logic is inverted. j.ExpiresAt.After(time.Now()) returns true when the token is still valid (expires in the future), but the method name suggests it should return true when expired. This will cause valid tokens to be rejected and expired tokens to be accepted.

Additionally, there's no nil check for j.ExpiresAt, which could cause a panic if the expiration time is not set.

Apply this diff to fix both issues:

 func (j *JWT) IsExpired() bool {
-	return j.ExpiresAt.After(time.Now())
+	if j.ExpiresAt == nil {
+		return true
+	}
+	return time.Now().After(j.ExpiresAt.Time)
 }
in priority.go
Missing validation for Impact allowed values.

The Impact field accepts any integer, but per the comment and ImpactToDecimal function, only specific values (25, 50, 100, 200, 300) are meaningful. Similarly, Confidence allows 50-100 but only 50, 80, 100 appear to be valid discrete values.

Consider adding custom validation or an oneof constraint if the binding framework supports it:

 type RICEParams struct {
     Reach      int `json:"reach" binding:"required,min=1"`       // Users per quarter
-	Impact     int `json:"impact" binding:"required"`            // 25, 50, 100, 200, 300
-	Confidence int `json:"confidence" binding:"required,min=50,max=100"` // 50, 80, 100
+	Impact     int `json:"impact" binding:"required,oneof=25 50 100 200 300"` // 25, 50, 100, 200, 300
+	Confidence int `json:"confidence" binding:"required,oneof=50 80 100"`     // 50, 80, 100
     Effort     int `json:"effort" binding:"required,min=1"`      // Person-weeks
 }

in user.go
Fix type inconsistency for the Role field.

The Role field is declared as string, but OrganizationMember.Role uses the OrganizationRole type. This inconsistency breaks type safety and could allow invalid role values to be assigned.

Apply this diff to use the proper type:

 // UserWithOrg represents a user with their organization context
 type UserWithOrg struct {
     User         *User         `json:"user"`
     Organization *Organization `json:"organization"`
-	Role         string        `json:"role"`
+	Role         OrganizationRole `json:"role"`
 }
in template_repository.go
Handle uuid.Parse errors to prevent silent failures.

The error from uuid.Parse is ignored, which could lead to silent failures if the database contains invalid UUID strings. This pattern appears in multiple methods (GetByID, ListByOrganization, ListByType, GetDefault).

Apply this diff to handle the error properly:

 if orgID.Valid {
-	oid, _ := uuid.Parse(orgID.String)
+	oid, err := uuid.Parse(orgID.String)
+	if err != nil {
+		return nil, err
+	}
     template.OrganizationID = &oid
 }
Apply similar fixes at:

Line 207 in ListByOrganization
Line 257 in ListByType
Line 307 in GetDefault
in the issue.go
Add nil-check for parent project and consider API call optimization.

Issue 1: Potential nil pointer dereference Line 454 accesses parent.Fields.Project.Key without checking if Project is nil. While unlikely, if the parent issue response doesn't include the project field, this will panic.

Apply this diff to add a nil check:

     parent, err := c.GetIssue(ctx, parentKey, nil)
     if err != nil {
         return nil, err
     }
+	if parent.Fields.Project == nil {
+		return nil, fmt.Errorf("parent issue %s has no project information", parentKey)
+	}

     req := &CreateIssueRequest{
Issue 2: Optimization opportunity The method makes an extra API call (line 447) solely to retrieve the parent's project key. Consider accepting projectKey as a parameter to avoid this overhead when the caller already knows it.


in the aes.go
Add key validation and copy the key defensively.

The constructor has two security concerns:

Missing key length validation: AES requires keys of exactly 16, 24, or 32 bytes (for AES-128, AES-192, or AES-256). Invalid key lengths will cause errors during encryption/decryption rather than at construction time, leading to delayed failure.

Key not copied: Storing the slice directly allows external code to modify the key after construction, which is a security risk.

Apply this diff to validate the key and store a defensive copy:

-func NewAESCrypt(secretKey []byte) *AESCrypt {
-	return &AESCrypt{secretKey: secretKey}
+func NewAESCrypt(secretKey []byte) (*AESCrypt, error) {
+	if len(secretKey) != 16 && len(secretKey) != 24 && len(secretKey) != 32 {
+		return nil, fmt.Errorf("AESCrypt: invalid key length %d, must be 16, 24, or 32 bytes", len(secretKey))
+	}
+
+	keyCopy := make([]byte, len(secretKey))
+	copy(keyCopy, secretKey)
+
+	return &AESCrypt{secretKey: keyCopy}, nil
 }

in the summarizer.go
Missing output schema for NotesToPRDPrompt.

Unlike MeetingSummarizerPrompt and DiscoveryInsightsPrompt, this prompt lacks a corresponding output schema variable. This breaks the established pattern and may cause issues if the consuming code expects all JSON-requesting prompts to have validation schemas.

Consider adding a NotesToPRDOutputSchema variable that defines the expected structure for the six sections mentioned (Problems Discussed, User Mentions, Feature Ideas, Success Criteria Mentioned, Concerns/Risks, Scope Discussions) along with confidence levels. Would you like me to generate the schema definition?

in the rice.go
Clarify confidence validation rules.

The validation allows any integer value between 50 and 100 (inclusive), but the ConfidenceLevel constants (lines 124-128) define only three valid values: 50, 80, and 100. This inconsistency could confuse API users about which confidence values are acceptable.

Consider either:

Restricting validation to only allow 50, 80, and 100, OR
Documenting that confidence accepts any value 50-100, making the ConfidenceLevel constants just convenient presets
If you want to restrict to the three defined levels, apply this diff:

-	if params.Confidence < 50 || params.Confidence > 100 {
+	validConfidence := map[int]bool{50: true, 80: true, 100: true}
+	if !validConfidence[params.Confidence] {
         return ErrInvalidConfidence
     }
