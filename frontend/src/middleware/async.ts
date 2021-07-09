import {Handler} from 'express';

export default function asyncWrap(fn: Handler): Handler {
    return (req, res, next): Promise<any> => {
        return Promise.resolve(fn(req, res, next)).catch(next);
    };
}
