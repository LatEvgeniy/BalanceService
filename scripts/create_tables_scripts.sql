CREATE TABLE public.users (
	id uuid NOT NULL,
	name varchar(50) NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id)
);

CREATE TABLE public.currencies (
	id uuid NOT NULL,
	"name" varchar(3) NOT NULL,
	CONSTRAINT currencies_pk PRIMARY KEY (id)
);

CREATE TABLE public.balances (
	userid uuid NOT NULL,
	currencyid uuid NOT NULL,
	balance numeric NOT NULL,
	lockedbalance numeric NOT NULL,
	CONSTRAINT balances_fk FOREIGN KEY (userid) REFERENCES public.users(id),
	CONSTRAINT balances_fk_1 FOREIGN KEY (currencyid) REFERENCES public.currencies(id)
);