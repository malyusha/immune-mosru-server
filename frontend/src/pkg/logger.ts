import winston, { format, Logger as WinstonLogger, transports } from 'winston';

export type Logger = WinstonLogger

let logger;

export function newLogger(level: string = 'info', meta: any = {}): Logger {
    const loggerFormat = format.combine(
        format.errors(),
        format.timestamp(),
        format.metadata({ fillWith: ['package', 'application', 'version', 'variant'] }),
        format.json(),
        format.timestamp({
            format: 'DD.MM.YYYY HH:mm:ss.SSS Z'
        }),
    );

    const productionConsole = new transports.Console({
        level: level,
        handleExceptions: true,
        format: format.combine(
            format.splat(),
            format.json(),
            format.errors(),
        ),
    });

    const devConsole = new transports.Console({
        handleExceptions: true,
        level: level,
        format: format.combine(
            format.colorize(),
            format.splat(),
            format.simple(),
        )
    });

    logger = winston.createLogger({
        level,
        format: loggerFormat,
        defaultMeta: { ...meta },
        transports: [
            process.env.NODE_ENV === 'production' ? productionConsole : devConsole
        ]
    });

    winston.add(logger);
    return logger;
}

if (!logger) {
    logger = newLogger();
}

export default logger;
