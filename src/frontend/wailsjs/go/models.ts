export namespace api {
	
	export class AnalyzeRequest {
	    prompt: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.model = source["model"];
	    }
	}
	export class StructuredDataBlock {
	    keys: string[];
	    values: string[];
	    raw: string[];
	    json: string;
	    diff: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new StructuredDataBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keys = source["keys"];
	        this.values = source["values"];
	        this.raw = source["raw"];
	        this.json = source["json"];
	        this.diff = this.convertValues(source["diff"], DiffLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DiffLine {
	    marker: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marker = source["marker"];
	        this.content = source["content"];
	    }
	}
	export class DuplicateGroup {
	    lines: string[];
	    keep: string;
	    confidence: number;
	    diff: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new DuplicateGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lines = source["lines"];
	        this.keep = source["keep"];
	        this.confidence = source["confidence"];
	        this.diff = this.convertValues(source["diff"], DiffLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SectionBreakdown {
	    system: number;
	    examples: number;
	    history: number;
	    question: number;
	
	    static createFrom(source: any = {}) {
	        return new SectionBreakdown(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.system = source["system"];
	        this.examples = source["examples"];
	        this.history = source["history"];
	        this.question = source["question"];
	    }
	}
	export class AnalyzeResponse {
	    inputTokens: number;
	    sections: SectionBreakdown;
	    duplicates: DuplicateGroup[];
	    nearDuplicates: DuplicateGroup[];
	    duplicateExamples: DuplicateGroup[];
	    nearDuplicateExamples: DuplicateGroup[];
	    structuredData: StructuredDataBlock[];
	    suggestions: string[];
	    potentialSavings: number;
	    model: string;
	    costKnown: boolean;
	    estimatedCostUsd: number;
	    estimatedSavingsUsd: number;
	    score: number;
	    issues: string[];
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputTokens = source["inputTokens"];
	        this.sections = this.convertValues(source["sections"], SectionBreakdown);
	        this.duplicates = this.convertValues(source["duplicates"], DuplicateGroup);
	        this.nearDuplicates = this.convertValues(source["nearDuplicates"], DuplicateGroup);
	        this.duplicateExamples = this.convertValues(source["duplicateExamples"], DuplicateGroup);
	        this.nearDuplicateExamples = this.convertValues(source["nearDuplicateExamples"], DuplicateGroup);
	        this.structuredData = this.convertValues(source["structuredData"], StructuredDataBlock);
	        this.suggestions = source["suggestions"];
	        this.potentialSavings = source["potentialSavings"];
	        this.model = source["model"];
	        this.costKnown = source["costKnown"];
	        this.estimatedCostUsd = source["estimatedCostUsd"];
	        this.estimatedSavingsUsd = source["estimatedSavingsUsd"];
	        this.score = source["score"];
	        this.issues = source["issues"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ApplyRequest {
	    prompt: string;
	    approveNearDuplicates: boolean;
	    approveJsonConversion: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.approveNearDuplicates = source["approveNearDuplicates"];
	        this.approveJsonConversion = source["approveJsonConversion"];
	    }
	}
	export class ApplyResponse {
	    optimized: string;
	    fullDiff: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new ApplyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.optimized = source["optimized"];
	        this.fullDiff = this.convertValues(source["fullDiff"], DiffLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	

}

export namespace store {
	
	export class ModelStats {
	    model: string;
	    requestCount: number;
	    tokensSaved: number;
	    savingsUsd: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.requestCount = source["requestCount"];
	        this.tokensSaved = source["tokensSaved"];
	        this.savingsUsd = source["savingsUsd"];
	    }
	}
	export class DashboardStats {
	    requestCount: number;
	    tokensSaved: number;
	    savingsUsd: number;
	    byModel: ModelStats[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestCount = source["requestCount"];
	        this.tokensSaved = source["tokensSaved"];
	        this.savingsUsd = source["savingsUsd"];
	        this.byModel = this.convertValues(source["byModel"], ModelStats);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Request {
	    id: number;
	    createdAt: string;
	    source: string;
	    model: string;
	    inputTokens: number;
	    optimizedTokens: number;
	    savingsPct: number;
	    costUsd: number;
	    savingsUsd: number;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.source = source["source"];
	        this.model = source["model"];
	        this.inputTokens = source["inputTokens"];
	        this.optimizedTokens = source["optimizedTokens"];
	        this.savingsPct = source["savingsPct"];
	        this.costUsd = source["costUsd"];
	        this.savingsUsd = source["savingsUsd"];
	    }
	}

}

export namespace proxy {
	
	export class Status {
	    running: boolean;
	    addr: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.addr = source["addr"];
	    }
	}

}

