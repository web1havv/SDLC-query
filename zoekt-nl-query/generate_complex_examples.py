#!/usr/bin/env python3
"""
Generate 200 more complex and advanced NL-to-Zoekt query pairs.
These are more sophisticated queries covering edge cases, complex patterns,
and real-world scenarios.
"""

import json

def generate_complex_examples():
    """Generate 200 complex NL-to-Zoekt query examples"""
    examples = []
    
    # Complex boolean queries
    complex_bool = [
        ("find Python functions but exclude test files", "lang:python content:def  -file:test"),
        ("search for JavaScript classes but not in node_modules", "lang:javascript content:class  -file:node_modules"),
        ("find Go functions excluding vendor and test directories", "lang:go content:func  -file:vendor -file:test"),
        ("search for Python imports but not in __pycache__", "lang:python content:import  -file:__pycache__"),
        ("find error handling in Python or JavaScript", "(lang:python or lang:javascript) content:error"),
        ("search for authentication in backend but not in tests", "repo:backend content:auth -file:test"),
        ("find database queries in Python or Go files", "(lang:python or lang:go) content:database"),
        ("search for API endpoints excluding mock files", "content:endpoint -file:mock"),
        ("find logging statements in Python but not in config files", "lang:python content:log -file:config"),
        ("search for async functions in JavaScript or TypeScript", "(lang:javascript or lang:typescript) content:async"),
    ]
    
    # Multi-condition file searches
    multi_file = [
        ("find configure_gemini in services.py or config.py", "file:services.py configure_gemini or file:config.py configure_gemini"),
        ("search for database connection in config or settings files", "(file:config or file:settings) content:database"),
        ("find API routes in routes.py or api.py", "(file:routes.py or file:api.py) content:route"),
        ("search for models in models.py or schema.py", "(file:models.py or file:schema.py)"),
        ("find middleware in middleware.py or utils.py", "(file:middleware.py or file:utils.py) content:middleware"),
        ("search for handlers in handler.py or controller.py", "(file:handler.py or file:controller.py)"),
        ("find validators in validator.py or validation.py", "(file:validator.py or file:validation.py)"),
        ("search for serializers in serializer.py or serializers.py", "(file:serializer.py or file:serializers.py)"),
    ]
    
    # Advanced regex patterns
    regex_advanced = [
        ("find email validation patterns", "content:/email.*@.*\\./"),
        ("search for UUID patterns in code", "content:/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/"),
        ("find date format patterns", "content:/\\d{4}-\\d{2}-\\d{2}/"),
        ("search for URL patterns", "content:/https?:\\/\\//"),
        ("find IP address patterns", "content:/\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}/"),
        ("search for credit card patterns", "content:/\\d{4}[\\s-]?\\d{4}[\\s-]?\\d{4}[\\s-]?\\d{4}/"),
        ("find phone number patterns", "content:/\\(?\\d{3}\\)?[-.]?\\d{3}[-.]?\\d{4}/"),
        ("search for hex color codes", "content:/#[0-9a-fA-F]{6}/"),
    ]
    
    # Nested repository and language combinations
    repo_lang_combo = [
        ("find Python files in backend repository", "repo:backend lang:python"),
        ("search for JavaScript in frontend but exclude tests", "repo:frontend lang:javascript -file:test"),
        ("find Go handlers in API repository", "repo:api lang:go content:handler"),
        ("search for TypeScript components in frontend", "repo:frontend lang:typescript content:component"),
        ("find Python services in backend excluding migrations", "repo:backend lang:python content:service -file:migration"),
        ("search for JavaScript utilities in shared repository", "repo:shared lang:javascript content:util"),
        ("find Go tests in test repository", "repo:test lang:go file:test"),
        ("search for Python scripts in scripts repository", "repo:scripts lang:python"),
    ]
    
    # Complex symbol searches
    complex_symbols = [
        ("find getAllBlogs function definition", "sym:getAllBlogs"),
        ("search for createUser method", "sym:createUser"),
        ("find processPayment function", "sym:processPayment"),
        ("search for validateToken method", "sym:validateToken"),
        ("find handleRequest function", "sym:handleRequest"),
        ("search for parseData method", "sym:parseData"),
        ("find formatOutput function", "sym:formatOutput"),
        ("search for authenticateUser method", "sym:authenticateUser"),
        ("find generateReport function", "sym:generateReport"),
        ("search for sendEmail method", "sym:sendEmail"),
    ]
    
    # Advanced comment searches
    comment_advanced = [
        ("find TODO comments in Python but not in tests", "lang:python content:/#.*TODO/ -file:test"),
        ("search for FIXME in JavaScript excluding node_modules", "lang:javascript content:/\\/\\/.*FIXME/ -file:node_modules"),
        ("find HACK comments in Go code", "lang:go content:/\\/\\/.*HACK/"),
        ("search for NOTE comments in any language", "content:/\\/\\/.*NOTE/ or content:/#.*NOTE/"),
        ("find deprecated warnings in comments", "content:/\\/\\/.*deprecated/ or content:/#.*deprecated/"),
        ("search for XXX markers in code comments", "content:/\\/\\/.*XXX/ or content:/#.*XXX/"),
        ("find BUG comments in Python files", "lang:python content:/#.*BUG/"),
        ("find OPTIMIZE comments in JavaScript", "lang:javascript content:/\\/\\/.*OPTIMIZE/"),
    ]
    
    # Complex content searches with multiple terms
    multi_content = [
        ("find where we use both database and cache", "content:database content:cache"),
        ("search for authentication and authorization together", "content:auth content:authorize"),
        ("find error handling with logging", "content:error content:log"),
        ("search for API calls with retry logic", "content:API content:retry"),
        ("find database transactions with rollback", "content:transaction content:rollback"),
        ("search for encryption and decryption functions", "content:encrypt content:decrypt"),
        ("find validation and sanitization together", "content:validate content:sanitize"),
        ("search for rate limiting and throttling", "content:rate.*limit content:throttle"),
    ]
    
    # Language-specific advanced patterns
    lang_advanced = [
        ("find Python decorators with parameters", "lang:python content:@.*("),
        ("search for Python context managers", "lang:python content:with "),
        ("find Python type hints with return types", "lang:python content:->"),
        ("search for Python list comprehensions", "lang:python content:for.*in.*]"),
        ("find Python generators", "lang:python content:yield"),
        ("search for JavaScript destructuring assignments", "lang:javascript content:const.*{"),
        ("find JavaScript template literals", "lang:javascript content:`"),
        ("search for JavaScript spread operators", "lang:javascript content:..."),
        ("find JavaScript optional chaining", "lang:javascript content:\\?\\."),
        ("search for Go error handling patterns", "lang:go content:if.*err"),
        ("find Go goroutines", "lang:go content:go "),
        ("search for Go channels", "lang:go content:chan"),
        ("find Go interfaces", "lang:go content:type.*interface"),
        ("search for Go struct definitions", "lang:go content:type.*struct"),
    ]
    
    # Real-world scenario queries
    real_world = [
        ("find where we handle payment processing", "content:payment content:process"),
        ("search for user authentication flow", "content:user content:auth"),
        ("find API rate limiting implementation", "content:rate.*limit"),
        ("search for database migration scripts", "file:migration"),
        ("find where we validate user input", "content:validate content:input"),
        ("search for session management code", "content:session"),
        ("find where we handle file uploads", "content:upload content:file"),
        ("search for email sending functionality", "content:email content:send"),
        ("find where we cache API responses", "content:cache content:API"),
        ("search for password hashing implementation", "content:password content:hash"),
        ("find where we generate JWT tokens", "content:JWT content:token"),
        ("search for OAuth implementation", "content:OAuth"),
        ("find where we handle CORS", "content:CORS"),
        ("search for error logging and monitoring", "content:error content:log"),
        ("find where we implement retry logic", "content:retry"),
    ]
    
    # Edge cases and special patterns
    edge_cases = [
        ("find functions with empty parameters", "content:def ( )"),
        ("search for functions with multiple parameters", "content:def .*,"),
        ("find async functions with await", "content:async content:await"),
        ("search for functions that return None", "content:return None"),
        ("find functions that raise exceptions", "content:raise"),
        ("search for functions with type annotations", "content:->"),
        ("find functions with decorators", "content:@ content:def"),
        ("search for functions with docstrings", "content:\"\"\""),
        ("find functions with comments above them", "content:# content:def"),
        ("search for functions with try-except blocks", "content:try content:except"),
    ]
    
    # File extension and pattern searches
    file_patterns = [
        ("find all .py files in src directory", "file:src file:.py"),
        ("search for .js files excluding node_modules", "file:.js -file:node_modules"),
        ("find .go files in cmd directory", "file:cmd file:.go"),
        ("search for .tsx files in components", "file:components file:.tsx"),
        ("find .json configuration files", "file:.json file:config"),
        ("search for .yaml or .yml files", "file:.yaml or file:.yml"),
        ("find .md documentation files", "file:.md"),
        ("search for .env environment files", "file:.env"),
    ]
    
    # Complex exclusion patterns
    exclusions = [
        ("find Python code excluding tests and migrations", "lang:python -file:test -file:migration"),
        ("search for JavaScript excluding node_modules and dist", "lang:javascript -file:node_modules -file:dist"),
        ("find Go code excluding vendor and generated files", "lang:go -file:vendor -file:generated"),
        ("search for TypeScript excluding build and node_modules", "lang:typescript -file:build -file:node_modules"),
        ("find all code excluding test and mock files", "content:def  -file:test -file:mock"),
        ("search for functions excluding private methods", "content:def  -content:__"),
        ("find public methods excluding getters and setters", "content:def  -content:get_ -content:set_"),
    ]
    
    # Count and aggregation queries
    count_queries = [
        ("how many Python functions are there", "lang:python content:def "),
        ("count JavaScript classes", "lang:javascript content:class "),
        ("how many Go packages", "lang:go content:package "),
        ("count TypeScript interfaces", "lang:typescript content:interface"),
        ("how many test files", "file:test"),
        ("count TODO comments", "content:/TODO/"),
        ("how many error handlers", "content:error"),
        ("count API endpoints", "content:endpoint"),
    ]
    
    # Specific framework patterns
    framework_patterns = [
        ("find React component definitions", "lang:javascript content:export.*function.*Component"),
        ("search for Express.js routes", "lang:javascript content:app\\.(get|post|put|delete)"),
        ("find Django models", "lang:python content:class.*models\\.Model"),
        ("search for Flask routes", "lang:python content:@app\\.route"),
        ("find FastAPI endpoints", "lang:python content:@app\\.(get|post|put|delete)"),
        ("search for Next.js API routes", "file:api lang:javascript"),
        ("find Vue component definitions", "lang:javascript content:export.*default"),
        ("search for Angular services", "lang:typescript content:@Injectable"),
    ]
    
    # Security-related queries
    security = [
        ("find where we sanitize user input", "content:sanitize content:input"),
        ("search for SQL injection prevention", "content:SQL content:injection"),
        ("find password hashing implementations", "content:password content:hash"),
        ("search for XSS prevention", "content:XSS content:escape"),
        ("find CSRF token handling", "content:CSRF content:token"),
        ("search for authentication middleware", "content:auth content:middleware"),
        ("find authorization checks", "content:authorize content:check"),
        ("search for encryption usage", "content:encrypt"),
    ]
    
    # Performance-related queries
    performance = [
        ("find where we use caching", "content:cache"),
        ("search for database query optimization", "content:query content:optimize"),
        ("find lazy loading implementations", "content:lazy content:load"),
        ("search for pagination logic", "content:pagination"),
        ("find where we batch operations", "content:batch"),
        ("search for connection pooling", "content:pool content:connection"),
        ("find where we use indexes", "content:index"),
        ("search for compression usage", "content:compress"),
    ]
    
    # Testing-related queries
    testing = [
        ("find unit test files", "file:test content:test"),
        ("search for test fixtures", "content:fixture"),
        ("find mock implementations", "content:mock"),
        ("search for test assertions", "content:assert"),
        ("find integration tests", "file:integration content:test"),
        ("search for test setup and teardown", "content:setup content:teardown"),
        ("find test data factories", "content:factory content:test"),
    ]
    
    # Combine all examples
    all_examples = (
        complex_bool + multi_file + regex_advanced + repo_lang_combo +
        complex_symbols + comment_advanced + multi_content + lang_advanced +
        real_world + edge_cases + file_patterns + exclusions +
        count_queries + framework_patterns + security + performance + testing
    )
    
    # Convert to JSON format
    for instruction, output in all_examples:
        examples.append({
            "instruction": instruction,
            "output": output
        })
    
    # Add more variations to reach exactly 200
    additional_complex = [
        # More complex boolean combinations
        ("find Python or JavaScript functions but not in tests", "(lang:python or lang:javascript) content:def  -file:test"),
        ("search for Go or Python error handling", "(lang:go or lang:python) content:error"),
        ("find TypeScript or JavaScript components", "(lang:typescript or lang:javascript) content:component"),
        
        # More file combinations
        ("find configuration in .env or config files", "(file:.env or file:config)"),
        ("search for tests in test or spec files", "(file:test or file:spec)"),
        ("find documentation in .md or .txt files", "(file:.md or file:.txt)"),
        
        # More regex patterns
        ("find version numbers", "content:/\\d+\\.\\d+\\.\\d+/"),
        ("search for semver patterns", "content:/\\d+\\.\\d+\\.\\d+/"),
        ("find timestamp patterns", "content:/\\d{10,13}/"),
        
        # More symbol combinations
        ("find where we call getAllBlogs", "content:getAllBlogs"),
        ("search for usage of createUser", "content:createUser"),
        ("find references to processPayment", "content:processPayment"),
        
        # More language-specific
        ("find Python async generators", "lang:python content:async content:yield"),
        ("search for JavaScript async/await patterns", "lang:javascript content:async content:await"),
        ("find Go concurrent patterns", "lang:go content:go content:chan"),
        
        # More real-world scenarios
        ("find where we handle file downloads", "content:download content:file"),
        ("search for image processing code", "content:image content:process"),
        ("find where we generate PDFs", "content:PDF content:generate"),
        ("search for CSV parsing", "content:CSV content:parse"),
        ("find where we handle webhooks", "content:webhook"),
        ("search for queue processing", "content:queue content:process"),
        
        # More edge cases
        ("find functions with default parameters", "content:def .*="),
        ("search for functions with keyword arguments", "content:def .*\\*\\*"),
        ("find functions with variable arguments", "content:def .*\\*"),
        ("search for lambda functions", "content:lambda"),
        ("find arrow functions", "content:=>"),
        
        # More framework patterns
        ("find Redux actions", "content:action content:type"),
        ("search for GraphQL resolvers", "content:resolver"),
        ("find REST API controllers", "content:controller content:API"),
        ("search for gRPC service definitions", "content:service content:rpc"),
        
        # More security patterns
        ("find input validation", "content:validate content:input"),
        ("search for output encoding", "content:encode content:output"),
        ("find secret management", "content:secret content:key"),
        ("search for token validation", "content:token content:validate"),
        
        # More performance patterns
        ("find database indexing", "content:index content:database"),
        ("search for query optimization", "content:query content:optimize"),
        ("find memory management", "content:memory content:manage"),
        ("search for resource cleanup", "content:cleanup content:close"),
    ]
    
    examples.extend(additional_complex)
    
    # Add final 9 examples to reach exactly 200
    final_examples = [
        ("find where we handle file compression", "content:compress content:file"),
        ("search for data serialization", "content:serialize"),
        ("find where we parse JSON responses", "content:JSON content:parse"),
        ("search for XML processing", "content:XML content:process"),
        ("find where we handle timeouts", "content:timeout"),
        ("search for circuit breaker patterns", "content:circuit.*breaker"),
        ("find where we implement backoff strategies", "content:backoff"),
        ("search for health check endpoints", "content:health content:check"),
        ("find where we handle graceful shutdown", "content:shutdown content:graceful"),
    ]
    
    examples.extend(final_examples)
    
    # Ensure we have exactly 200
    examples = examples[:200]
    
    return examples

def save_examples(examples, filename="zoekt_examples_complex.jsonl"):
    """Save examples to JSONL format"""
    with open(filename, 'w') as f:
        for example in examples:
            f.write(json.dumps(example) + '\n')
    print(f"✅ Saved {len(examples)} complex examples to {filename}")

if __name__ == "__main__":
    print("🚀 Generating 200 complex NL-to-Zoekt query examples...")
    print("   (Advanced patterns, edge cases, real-world scenarios)")
    print("")
    
    examples = generate_complex_examples()
    
    print(f"✅ Generated {len(examples)} complex examples")
    save_examples(examples)
    
    print("")
    print("📊 Example breakdown:")
    complex_count = sum(1 for e in examples if isinstance(e, dict) and ('but' in e.get('instruction', '').lower() or 'excluding' in e.get('instruction', '').lower() or 'or' in e.get('instruction', '').lower()))
    regex_count = sum(1 for e in examples if isinstance(e, dict) and '/' in e.get('output', ''))
    multi_count = sum(1 for e in examples if isinstance(e, dict) and 'content:' in e.get('output', '') and e.get('output', '').count('content:') > 1)
    print(f"   - Complex boolean queries: {complex_count}")
    print(f"   - Regex patterns: {regex_count}")
    print(f"   - Multi-condition queries: {multi_count}")
    print("")
    print("🎯 Next step: Merge with existing examples and rebuild index")

