#!/usr/bin/env node
'use strict';

const fs = require('fs');
const vm = require('vm');
const path = require('path');

if (process.argv.length < 4) {
  console.error('Usage: build_client_routes.js <router.js path> <output json path>');
  process.exit(1);
}

const routerPath = process.argv[2];
const outputPath = process.argv[3];

const source = fs.readFileSync(routerPath, 'utf8');

const cleaned = source
  .replace(/^\s*import .*$/mg, '')
  .replace(/^\s*export default /mg, '');

const prelude = `
const __routes = [];
const __pathStack = [];
const __nameStack = [];

function __joinPath(parent, segment) {
  let base = parent || '';
  if (base === '/') {
    base = '';
  }
  let seg = segment || '';
  if (seg === '/' || seg === '') {
    seg = '';
  }
  if (seg.startsWith('/')) {
    seg = seg.slice(1);
  }
  let combined = '';
  if (base && seg) {
    combined = base + '/' + seg;
  } else if (base) {
    combined = base;
  } else if (seg) {
    combined = seg;
  }
  if (!combined.startsWith('/')) {
    combined = '/' + combined;
  }
  if (combined === '') {
    combined = '/';
  }
  return combined;
}

function __extractParams(path) {
  const params = [];
  for (const part of path.split('/')) {
    if (part.startsWith(':') || part.startsWith('*')) {
      const name = part.slice(1).trim();
      if (name) params.push(name);
    }
  }
  return params;
}

function __route(name, options, callback) {
  let opts = options;
  let cb = callback;
  if (typeof opts === 'function') {
    cb = opts;
    opts = null;
  }
  const pathOpt = opts && opts.path ? String(opts.path) : String(name);
  const resetNamespace = opts && opts.resetNamespace === true;

  const parentPath = __pathStack.length ? __pathStack[__pathStack.length - 1] : '';
  const parentNames = resetNamespace ? [] : (__nameStack.length ? __nameStack[__nameStack.length - 1] : []);
  const currentNames = parentNames.concat([String(name)]);

  const fullPath = __joinPath(parentPath, pathOpt);
  const key = currentNames.join('.');
  const params = __extractParams(fullPath);

  __routes.push({
    key,
    name: String(name),
    path: fullPath,
    params,
  });

  if (typeof cb === 'function') {
    __pathStack.push(fullPath);
    __nameStack.push(currentNames);
    cb.call(__routeRecorder);
    __nameStack.pop();
    __pathStack.pop();
  }
}

const __routeRecorder = { route: __route };
class EmberRouter {
  static map(cb) {
    cb.call(__routeRecorder);
  }
}
const config = { locationType: 'hash', rootURL: '/' };
`;

const postlude = `
this.__routes = __routes;
`;

const script = prelude + '\n' + cleaned + '\n' + postlude;

const context = {
  console,
  require,
  module: {},
  exports: {},
};
vm.createContext(context);
vm.runInContext(script, context, { filename: path.basename(routerPath) });

const routes = context.__routes || [];

function isActionRoute(name) {
  return name === 'new' || name === 'edit' || name === 'destroy';
}

function terminalInfo(route) {
  const segments = route.path.split('/').filter(Boolean);
  for (let i = segments.length - 1; i >= 1; i--) {
    const seg = segments[i];
    const prev = segments[i - 1];
    if ((seg.startsWith(':') || seg.startsWith('*')) && prev && !prev.startsWith(':') && !prev.startsWith('*')) {
      return { segment: prev, param: seg.slice(1) };
    }
  }
  return { segment: '', param: '' };
}

const catalog = routes.map(route => {
  const terminal = terminalInfo(route);
  return {
    key: route.key,
    path: route.path,
    params: route.params,
    action: isActionRoute(route.name),
    terminal_segment: terminal.segment,
    terminal_param: terminal.param,
  };
});

fs.writeFileSync(outputPath, JSON.stringify({ routes: catalog }, null, 2));
console.log(`Wrote ${catalog.length} routes to ${outputPath}`);
