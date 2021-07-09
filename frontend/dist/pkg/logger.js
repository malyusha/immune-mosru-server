"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.newLogger = void 0;
const winston_1 = __importStar(require("winston"));
let logger;
function newLogger(level = 'info', meta = {}) {
    const loggerFormat = winston_1.format.combine(winston_1.format.errors(), winston_1.format.timestamp(), winston_1.format.metadata({ fillWith: ['package', 'application', 'version', 'variant'] }), winston_1.format.json(), winston_1.format.timestamp({
        format: 'DD.MM.YYYY HH:mm:ss.SSS Z'
    }));
    const productionConsole = new winston_1.transports.Console({
        level: level,
        handleExceptions: true,
        format: winston_1.format.combine(winston_1.format.splat(), winston_1.format.json(), winston_1.format.errors()),
    });
    const devConsole = new winston_1.transports.Console({
        handleExceptions: true,
        level: level,
        format: winston_1.format.combine(winston_1.format.colorize(), winston_1.format.splat(), winston_1.format.simple())
    });
    logger = winston_1.default.createLogger({
        level,
        format: loggerFormat,
        defaultMeta: Object.assign({}, meta),
        transports: [
            process.env.NODE_ENV === 'production' ? productionConsole : devConsole
        ]
    });
    winston_1.default.add(logger);
    return logger;
}
exports.newLogger = newLogger;
if (!logger) {
    logger = newLogger();
}
exports.default = logger;
