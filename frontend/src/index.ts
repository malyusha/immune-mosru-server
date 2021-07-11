import 'module-alias/register';
import express from 'express';
import path from 'path';
import yargs from 'yargs';
import asyncWrap from './middleware/async';
import loadConfiguration from './config';
import {Config} from './types';
import CertsClient from './pkg/clients/certs';
import logger from "@app/pkg/logger";
import moment from 'moment';
import LRU from 'lru-cache';

async function main(): Promise<void> {
    const argv = yargs
        .option('config', {
            alias: 'c',
            demandOption: false,
            description: 'Configuration file path for server',
            type: 'string'
        }).argv;

    let config: Config;
    try {
        config = loadConfiguration(argv.config);
    } catch (e) {
        logger.error(`failed to load configuration: ${e}`);
        process.exit(1);
    }


    const app = express();
    const router = express.Router();

    app.set('view engine', 'pug');
    app.set('views', path.join(__dirname, 'pages'));
    app.use('/assets', express.static(path.join(__dirname, 'assets')));

    const certsClient = new CertsClient(config.clients.certs);
    moment.locale('ru');
    const cache = new LRU({
        max: 2000,
        maxAge: 1000 * 3600,
    });

    router.get('/qr', asyncWrap(async (req, res) => {
        const id = req.query.id;
        if (id === '' || typeof id !== 'string') {
            return res.render('qr', {})
        }

        let renderData = {
            data: {},
            vaccinated: false,
        };

        const cached = cache.get(id);
        if (cached) {
            renderData = cached;
        } else {
            logger.info(`handling qr code ${id}`);
            try {
                const {dateBirth, ...data} = await certsClient.getQRDataByCode(id);
                renderData.vaccinated = true;
                logger.info(`date birth ${dateBirth}`);
                renderData.data = {
                    dateBirth: moment(dateBirth, 'DD.MM.YYYY').format('DD MMMM'),
                    ...data,
                };

                cache.set(id, renderData);
            } catch (e) {
                logger.error('failed to load data by code from service', e);
            }
        }

        return res.render('qr', renderData);
    }));

    app.disable('x-powered-by');
    app.use('/', router);
    app.use((req, res) => {
        return res.status(404).send('Not found');
    });

    app.listen(config.http.port, () => {
        logger.info(`Running at ${config.http.port}`);
    });

    ['SIGINT', 'SIGTERM'].forEach((sig) => {
        process.on(sig, (code) => {
            logger.info(`received process code: ${code}`);
            process.exit(code);
        });
    });
}

process.on('unhandledRejection', (error) => {
    logger.error(`unhandledRejection: ${error}`, {
        original_error: error,
    });
    setTimeout(() => process.exit(), 3000);
});

process.on('uncaughtException', (error) => {
    logger.error(`uncaughtException: ${error.message}`, {
        original_error: error,
    });
    setTimeout(() => process.exit(), 3000);
});

main();
