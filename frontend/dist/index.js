"use strict";
var __rest = (this && this.__rest) || function (s, e) {
    var t = {};
    for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p) && e.indexOf(p) < 0)
        t[p] = s[p];
    if (s != null && typeof Object.getOwnPropertySymbols === "function")
        for (var i = 0, p = Object.getOwnPropertySymbols(s); i < p.length; i++) {
            if (e.indexOf(p[i]) < 0 && Object.prototype.propertyIsEnumerable.call(s, p[i]))
                t[p[i]] = s[p[i]];
        }
    return t;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
require("module-alias/register");
const express_1 = __importDefault(require("express"));
const path_1 = __importDefault(require("path"));
const yargs_1 = __importDefault(require("yargs"));
const async_1 = __importDefault(require("./middleware/async"));
const config_1 = __importDefault(require("./config"));
const certs_1 = __importDefault(require("./pkg/clients/certs"));
const logger_1 = __importDefault(require("@app/pkg/logger"));
const moment_1 = __importDefault(require("moment"));
moment_1.default.locale('ru');
async function main() {
    const argv = yargs_1.default
        .option('config', {
        alias: 'c',
        demandOption: false,
        description: 'Configuration file path for server',
        type: 'string'
    }).argv;
    let config;
    try {
        config = config_1.default(argv.config);
    }
    catch (e) {
        console.log(`failed to load configuration: ${e}`);
        process.exit(1);
    }
    const app = express_1.default();
    const router = express_1.default.Router();
    app.set('view engine', 'pug');
    app.set('views', path_1.default.join(__dirname, 'pages'));
    app.use('/assets', express_1.default.static(path_1.default.join(__dirname, 'assets')));
    const certsClient = new certs_1.default(config.clients.certs);
    router.get('/qr', async_1.default(async (req, res) => {
        const id = req.query.id;
        if (id === '' || typeof id !== 'string') {
            return res.render('qr', {});
        }
        let renderData = {
            data: {},
            vaccinated: false,
        };
        logger_1.default.info(`handling qr code ${id}`);
        try {
            const _a = await certsClient.getQRDataByCode(id), { dateBirth } = _a, data = __rest(_a, ["dateBirth"]);
            renderData.vaccinated = true;
            logger_1.default.info(`date birth ${dateBirth}`);
            renderData.data = Object.assign({ dateBirth: moment_1.default(dateBirth, 'DD.MM.YYYY').format('DD MMMM') }, data);
        }
        catch (e) {
            logger_1.default.error('failed to load data by code from service', e);
        }
        return res.render('qr', renderData);
    }));
    app.disable('x-powered-by');
    app.use('/', router);
    app.use((req, res) => {
        return res.status(404).send('Not found');
    });
    const port = process.env.APP_PORT || 8080;
    app.listen(port, () => {
        console.log(`Running at ${port}`);
    });
}
process.on('unhandledRejection', (error) => {
    logger_1.default.error(`unhandledRejection: ${error}`, {
        original_error: error,
    });
    setTimeout(() => process.exit(), 3000);
});
process.on('uncaughtException', (error) => {
    logger_1.default.error(`uncaughtException: ${error.message}`, {
        original_error: error,
    });
    setTimeout(() => process.exit(), 3000);
});
main();
