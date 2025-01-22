export namespace main {
	
	export class Result {
	    URL: string;
	    Size: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.URL = source["URL"];
	        this.Size = source["Size"];
	    }
	}

}

