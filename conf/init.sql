--  This file is part of the eliona project.
--  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists signify;

-- Should be editable by eliona frontend.
create table if not exists signify.configuration
(
    id               bigserial primary key,
    base_url         text    not null,
    service          text    not null,
    service_id       text    not null,
    service_secret   text    not null,
    app_key          text    not null,
    app_secret       text    not null,
    refresh_interval integer not null default 60,
    request_timeout  integer not null default 120,
    asset_filter     json,
    active           boolean          default false,
    enable           boolean          default false,
    project_ids      text[]
);

create table if not exists signify.asset
(
	id               bigserial primary key,
	kind             text      not null,
    uuid             text      not null,
    parent_uuid      text      null,
	configuration_id bigserial not null references signify.configuration(id),
	project_id       text      not null,
	global_asset_id  text      not null,
	asset_id         integer
);

-- Makes the new objects available for all other init steps
commit;
