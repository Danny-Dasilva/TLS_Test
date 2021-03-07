"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (Object.prototype.hasOwnProperty.call(b, p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
var __assign = (this && this.__assign) || function () {
    __assign = Object.assign || function(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];
            for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                t[p] = s[p];
        }
        return t;
    };
    return __assign.apply(this, arguments);
};
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __spreadArrays = (this && this.__spreadArrays) || function () {
    for (var s = 0, i = 0, il = arguments.length; i < il; i++) s += arguments[i].length;
    for (var r = Array(s), k = 0, i = 0; i < il; i++)
        for (var a = arguments[i], j = 0, jl = a.length; j < jl; j++, k++)
            r[k] = a[j];
    return r;
};
exports.__esModule = true;
var child_process_1 = require("child_process");
var path_1 = require("path");
var events_1 = require("events");
var ws_1 = require("ws");
var child;
var cleanExit = function (message) {
    if (message)
        console.log(message);
    child.kill();
    process.exit();
};
process.on('SIGINT', function () { return cleanExit(); });
process.on('SIGTERM', function () { return cleanExit(); });
var Golang = /** @class */ (function (_super) {
    __extends(Golang, _super);
    function Golang(port, debug) {
        var _this = _super.call(this) || this;
        _this.server = new ws_1.Server({ port: port });
        var executableFilename;
        if (process.platform == 'win32') {
            executableFilename = 'index.exe';
        }
        else if (process.platform == 'linux') {
            executableFilename = 'index';
        }
        else if (process.platform == 'darwin') {
            executableFilename = 'index-mac';
        }
        else {
            cleanExit(new Error('Operating system not supported'));
        }
        child = child_process_1.spawn(path_1["default"].join(__dirname, executableFilename), {
            env: { WS_PORT: port.toString() },
            shell: true,
            windowsHide: true
        });
        child.stderr.on('data', function (stderr) {
            if (stderr.toString().includes('Request_Id_On_The_Left')) {
                var splitRequestIdAndError = stderr.toString().split('Request_Id_On_The_Left');
                var requestId = splitRequestIdAndError[0], error = splitRequestIdAndError[1];
                _this.emit(requestId, { error: new Error(error) });
            }
            else {
                debug
                    ? cleanExit(new Error(stderr))
                    : cleanExit(new Error('Invalid JA3 hash. Exiting... (Golang wrapper exception)'));
            }
        });
        _this.server.on('connection', function (ws) {
            _this.emit('ready');
            ws.on('message', function (data) {
                var message = JSON.parse(data);
                _this.emit(message.RequestID, message.Response);
            });
        });
        return _this;
    }
    Golang.prototype.request = function (requestId, options) {
        __spreadArrays(this.server.clients)[0].send(JSON.stringify({ requestId: requestId, options: options }));
    };
    return Golang;
}(events_1.EventEmitter));
var initMyTls = function (initOptions) {
    if (initOptions === void 0) { initOptions = {}; }
    return __awaiter(void 0, void 0, void 0, function () {
        return __generator(this, function (_a) {
            return [2 /*return*/, new Promise(function (resolveReady) {
                    var port = initOptions.port, debug = initOptions.debug;
                    if (!port)
                        port = 9119;
                    if (!debug)
                        debug = false;
                    var instance = new Golang(port, debug);
                    instance.on('ready', function () {
                        var mytls = (function () {
                            var MyTls = function (url, options, method) {
                                if (method === void 0) { method = 'get'; }
                                return __awaiter(void 0, void 0, void 0, function () {
                                    return __generator(this, function (_a) {
                                        return [2 /*return*/, new Promise(function (resolveRequest, rejectRequest) {
                                                var requestId = "" + url + Math.floor(Date.now() * Math.random());
                                                if (!options.ja3)
                                                    options.ja3 =
                                                        '771,255-49195-49199-49196-49200-49171-49172-156-157-47-53,0-10-11-13,23-24,0';
                                                if (!options.body)
                                                    options.body = '';
                                                if (!options.proxy)
                                                    options.proxy = '';
                                                instance.request(requestId, __assign(__assign({ url: url }, options), { method: method }));
                                                instance.once(requestId, function (response) {
                                                    if (response.error)
                                                        return rejectRequest(response.error);
                                                    var status = response.Status, body = response.Body, headers = response.Headers;
                                                    if (headers['Set-Cookie'])
                                                        headers['Set-Cookie'] = headers['Set-Cookie'].split('/,/');
                                                    resolveRequest({
                                                        status: status,
                                                        body: body,
                                                        headers: headers
                                                    });
                                                });
                                            })];
                                    });
                                });
                            };
                            MyTls.head = function (url, options) {
                                return MyTls(url, options, 'head');
                            };
                            MyTls.get = function (url, options) {
                                return MyTls(url, options, 'get');
                            };
                            MyTls.post = function (url, options) {
                                return MyTls(url, options, 'post');
                            };
                            MyTls.put = function (url, options) {
                                return MyTls(url, options, 'put');
                            };
                            MyTls["delete"] = function (url, options) {
                                return MyTls(url, options, 'delete');
                            };
                            MyTls.trace = function (url, options) {
                                return MyTls(url, options, 'trace');
                            };
                            MyTls.options = function (url, options) {
                                return MyTls(url, options, 'options');
                            };
                            MyTls.connect = function (url, options) {
                                return MyTls(url, options, 'options');
                            };
                            MyTls.patch = function (url, options) {
                                return MyTls(url, options, 'patch');
                            };
                            return MyTls;
                        })();
                        resolveReady(mytls);
                    });
                })];
        });
    });
};
exports["default"] = initMyTls;
// CommonJS support for default export
module.exports = initMyTls;
module.exports["default"] = initMyTls;
module.exports.__esModule = true;
