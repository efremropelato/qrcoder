export namespace main {
	
	export class QRParams {
	    url: string;
	    size: number;
	    fgColor: string;
	    bgColor: string;
	    logoPath: string;
	    errorLevel: string;
	    format: string;
	    disableBorder: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QRParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.size = source["size"];
	        this.fgColor = source["fgColor"];
	        this.bgColor = source["bgColor"];
	        this.logoPath = source["logoPath"];
	        this.errorLevel = source["errorLevel"];
	        this.format = source["format"];
	        this.disableBorder = source["disableBorder"];
	    }
	}

}

