grammar ADL

entry Model:
    declarations+=Declaration*;

Declaration:
      Namespace
    | Vocabulary
    | Entity
    | Relation
    | Brain
    | Mind
    | Belief
    | Constitution
    | Policy
    | Api
    | Function
    | SurrealMlBinding;

Namespace:
    'namespace' name=QualifiedName;

Vocabulary:
    'vocabulary' name=ID '{'
        terms+=Term*
    '}';

Term:
    'term' name=ID ('means' meaning=STRING)? ('maps' 'to' mapping=QualifiedName)?;

Entity:
    'entity' name=ID ('extends' super=[Entity:QualifiedName])? '{'
        fields+=Field*
    '}';

Field:
    name=ID ':' type=TypeRef (required?='required')?;

TypeRef:
      primitive=PrimitiveType
    | ref=[Entity:QualifiedName];

PrimitiveType:
    'string' | 'int' | 'float' | 'bool' | 'datetime' | 'object' | 'array' | 'record';

Relation:
    'relation' name=ID source=[Entity:QualifiedName] '->' target=[Entity:QualifiedName] '{'
        fields+=Field*
    '}';

Brain:
    'brain' name=ID '{'
        owns+=Ownership*
    '}';

Ownership:
    'owns' area=RuntimeArea;

RuntimeArea:
      'data'
    | 'business_logic'
    | 'policy'
    | 'api'
    | 'permissions'
    | 'events'
    | 'live_queries'
    | 'memory'
    | 'beliefs'
    | 'decisions';

Mind:
    'mind' name=ID 'for' subject=[Entity:QualifiedName] '{'
        fields+=Field*
    '}';

Belief:
    'belief' name=ID '{'
        'subject' subject=[Entity:QualifiedName]
        'claim' claim=STRING
        ('confidence' confidence=NUMBER)?
        ('source' source=STRING)?
    '}';

Constitution:
    'constitution' name=ID '{'
        rules+=Rule*
    '}';

Rule:
    'rule' name=ID ':' statement=STRING;

Policy:
    'policy' name=ID '{'
        rules+=Rule*
    '}';

Api:
    'api' method=HttpMethod path=STRING 'returns' target=SurrealTarget;

HttpMethod:
    'get' | 'post' | 'put' | 'patch' | 'delete';

Function:
    'function' name=QualifiedName '(' (params+=Param (',' params+=Param)*)? ')' 'returns' type=TypeRef 'as' target=SurrealTarget;

Param:
    name=ID ':' type=TypeRef;

SurrealMlBinding:
    'surrealml' name=ID 'for' purpose=InferencePurpose 'input' input=[Entity:QualifiedName] 'output' output=[Entity:QualifiedName];

InferencePurpose:
      'classification'
    | 'scoring'
    | 'prediction'
    | 'ranking'
    | 'embedding';

SurrealTarget:
      'surrealdb'
    | 'surrealml';

QualifiedName:
    ID ('.' ID)*;

terminal NUMBER returns number: /[0-9]+(\.[0-9]+)?/;
terminal ID: /[_a-zA-Z][\w_]*/;
terminal STRING: /"[^"\\]*(\\.[^"\\]*)*"/;
hidden terminal WS: /\s+/;
hidden terminal ML_COMMENT: /\/\*[\s\S]*?\*\//;
hidden terminal SL_COMMENT: /\/\/[^\n\r]*/;
