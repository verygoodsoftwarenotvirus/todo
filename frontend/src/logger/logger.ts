export const enum LogLevel {
  Meta = 0,
  Info,
  Debug,
  Warning,
  Error,
  WTF,
}

const defaultTextColor: string = 'ghostwhite';
const defaultLogLevel: LogLevel = LogLevel.Debug;

const infoPrefix: string = '%c  INFO   ',
  infoBackgroundColor: string = 'darkslategray',
  infoTextColor: string = defaultTextColor;

const debugPrefix: string = '%c  DEBUG  ',
  debugBackgroundColor: string = 'rebeccapurple',
  debugTextColor: string = defaultTextColor;

const warningPrefix: string = '%c WARNING ',
  warningBackgroundColor: string = 'goldenrod',
  warningTextColor: string = defaultTextColor;

const errorPrefix: string = '%c  ERROR  ',
  errorBackgroundColor: string = 'firebrick',
  errorTextColor: string = defaultTextColor;

const wtfPrefix: string = '%c   WTF   ',
  wtfBackgroundColor: string = 'palegreen',
  wtfTextColor: string = 'rosybrown';

export class Logger {
  includeContextByDefault: boolean;
  private _level: LogLevel;
  private readonly context: Map<string, any>;
  private readonly debugOnlyContext: Map<string, any>;

  protected actualLogFunc: (
    prefix: string,
    style: string,
    content: string,
    context?: Map<string, any> | string,
  ) => void;

  constructor(
    level: LogLevel = defaultLogLevel,
    context: Map<string, any> | null = null,
  ) {
    this._level = level;
    this.includeContextByDefault = false;
    this.context = context ? new Map(context) : new Map();
    this.debugOnlyContext = new Map();

    this.actualLogFunc = console.log;
  }

  private static buildStyle(
    bgColor: string = 'black',
    textColor: string = 'white',
  ): string {
    return `background: ${bgColor}; color: ${textColor};`;
  }

  // this function is for vanity only
  static demo(): void {
    const l = new Logger().withValue('demo', 'true');
    l.level = LogLevel.WTF;

    l.info('this is a demo');
    l.debug('this is a demo');
    l.warning('this is a demo');
    l.error('this is a demo');
    l.WTF('this is a demo');
  }

  toggleContextInclusion(): void {
    this.includeContextByDefault = !this.includeContextByDefault;
  }

  get level(): LogLevel {
    return this._level;
  }

  set level(ll: LogLevel) {
    this._level = ll;
  }

  fetchPageContext(): void {
    const l: Location = window.location;

    this.context.set('currentURL', l.toString());
    this.context.set('query', l.search);
    this.context.set('path', l.pathname);
  }

  withValue(key: string, value: any): Logger {
    let ctx = new Map(this.context);
    ctx.set(key, value);
    return new Logger(this.level, ctx);
  }

  withValues(values: Map<string, any>): Logger {
    let ctx = new Map(this.context);

    for (const x of values.entries()) {
      ctx.set(x[0], x[1]);
    }

    return new Logger(this.level, ctx);
  }

  withDebugValue(key: string, value: any): Logger {
    return this.withValue(key, value);
  }

  info(s: string = '', includeCtx: boolean = false) {
    const prefix = infoPrefix;
    const style = Logger.buildStyle(infoBackgroundColor, infoTextColor);

    this.includeContextByDefault || includeCtx
      ? this.actualLogFunc(prefix, style, s, this.context)
      : this.actualLogFunc(prefix, style, s);
  }

  debug(s: string = '', includeCtx: boolean = true) {
    const prefix = debugPrefix;
    const style = Logger.buildStyle(debugBackgroundColor, debugTextColor);

    let ctx = new Map(this.context);
    for (const x of this.debugOnlyContext.entries()) {
      if (!ctx.has(x[0])) {
        ctx.set(x[0], x[1]);
      }
    }

    if (this.level >= LogLevel.Debug) {
      this.includeContextByDefault || includeCtx
        ? this.actualLogFunc(prefix, style, s, ctx)
        : this.actualLogFunc(prefix, style, s);
    }
  }

  warning(s: string = '', includeCtx: boolean = false) {
    const prefix = warningPrefix;
    const style = Logger.buildStyle(warningBackgroundColor, warningTextColor);

    if (this.level >= LogLevel.Warning) {
      this.includeContextByDefault || includeCtx
        ? this.actualLogFunc(prefix, style, s, this.context)
        : this.actualLogFunc(prefix, style, s);
    }
  }

  error(s: string = '', includeCtx: boolean = true) {
    const prefix = errorPrefix;
    const style = Logger.buildStyle(errorBackgroundColor, errorTextColor);

    if (this.level >= LogLevel.Error) {
      this.includeContextByDefault || includeCtx
        ? this.actualLogFunc(prefix, style, s, this.context)
        : this.actualLogFunc(prefix, style, s);
    }
  }

  WTF(s: string = '', includeCtx: boolean = true) {
    const prefix = wtfPrefix;
    const style = Logger.buildStyle(wtfBackgroundColor, wtfTextColor);

    if (this.level >= LogLevel.WTF) {
      this.includeContextByDefault || includeCtx
        ? this.actualLogFunc(prefix, style, s, this.context)
        : this.actualLogFunc(prefix, style, s);
    }
  }
}
