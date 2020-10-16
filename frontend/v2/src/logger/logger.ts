export const enum LogLevel {
    Info = 0,
    Debug,
    Warning,
    Error,
    WTF,
}

const infoBackgroundColor = "black";
const debugBackgroundColor = "blue";
const warningBackgroundColor = "orange";
const errorBackgroundColor = "red";
const wtfBackgroundColor = "magenta";

const infoTextColor = "white";
const debugTextColor = "white";
const warningTextColor = "white";
const errorTextColor = "white";
const wtfTextColor = "tomato";

const infoPrefix = "%c  INFO   ";
const debugPrefix = "%c  DEBUG  ";
const warningPrefix = "%c WARNING ";
const errorPrefix = "%c  ERROR  ";
const wtfPrefix = "%c   WTF   ";

export class Logger {
    includeContextByDefault: boolean
    level: LogLevel;
    context: Map<string, string>;

    constructor(level: LogLevel = LogLevel.Info, context: Map<string, string> | null = null) {
        this.level = level;
        this.includeContextByDefault = false;
        this.context = context ? new Map(context) : new Map();
    }

    private static buildStyle(bgColor: string = 'black', textColor: string = 'white'): string {
        return `background: ${bgColor}; color: ${textColor};`
    }

    toggleContextInclusion(): void {
        this.includeContextByDefault = !this.includeContextByDefault;
    }

    setLevel(level: LogLevel): void {
        this.level = level;
    }

    fetchPageContext(): void {
        const l: Location = window.location;

        this.context.set("currentURL", l.toString());
        this.context.set("query", l.search);
        this.context.set("path", l.pathname);
    }

    withValue(key: string, value: string): Logger {
        let ctx = new Map(this.context);
        ctx.set(key, value);
        return new Logger(this.level, ctx);
    }

    withValues(values: Map<string, string>): Logger {
        let ctx = new Map(this.context);

        for (const x of values.entries()) {
            ctx.set(x[0], x[1]);
        }

        return new Logger(this.level, ctx);
    }

    info(s: string = '', includeCtx: boolean = false): void {
        const prefix = infoPrefix;
        const style = Logger.buildStyle(infoBackgroundColor, infoTextColor);

        ( this.includeContextByDefault || includeCtx ) ?
            console.log(prefix, style, this.context, s) :
            console.log(prefix, style, s);
    }

    debug(s: string = '', includeCtx: boolean = false): void {
        const prefix = debugPrefix;
        const style = Logger.buildStyle(debugBackgroundColor, debugTextColor);

        if (this.level >= LogLevel.Debug) {
            ( this.includeContextByDefault || includeCtx ) ?
                console.debug(prefix, style, this.context, s) :
                console.debug(prefix, style, s);
        }
    }

    warning(s: string = '', includeCtx: boolean = false): void {
        const prefix = warningPrefix;
        const style = Logger.buildStyle(warningBackgroundColor, warningTextColor);

        if (this.level >= LogLevel.Warning) {
            ( this.includeContextByDefault || includeCtx ) ?
                console.log(prefix, style, this.context, s) :
                console.log(prefix, style, s);
        }
    }

    error(s: string = '', includeCtx: boolean = false): void {
        const prefix = errorPrefix;
        const style = Logger.buildStyle(errorBackgroundColor, errorTextColor);

        if (this.level >= LogLevel.Error) {
            ( this.includeContextByDefault || includeCtx ) ?
                console.log(prefix, style, this.context, s) :
                console.log(prefix, style, s);
        }
    }

    WTF(s: string = '', includeCtx: boolean = false): void {
        const prefix = wtfPrefix;
        const style = Logger.buildStyle(wtfBackgroundColor, wtfTextColor);

        if (this.level >= LogLevel.WTF) {
            ( this.includeContextByDefault || includeCtx ) ?
                console.log(prefix, style, this.context, s) :
                console.log(prefix, style, s);
        }
    }
}
