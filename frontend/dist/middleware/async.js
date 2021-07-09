"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
function asyncWrap(fn) {
    return (req, res, next) => {
        return Promise.resolve(fn(req, res, next)).catch(next);
    };
}
exports.default = asyncWrap;
