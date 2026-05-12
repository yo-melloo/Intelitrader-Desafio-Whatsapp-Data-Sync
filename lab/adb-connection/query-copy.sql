-- O WhatsApp não salva o nome dos contatos na mesma tabela que salva os textos das mensagens;
-- Pior que isso, ele salva em outro banco de dados.
-- Para nossa modelagem de dados, pensando na camada de consumo (API -> Front-End), os dados trazidos são:
--  ID, Contexto, Nome_Conversa, Remetente, Conteudo, e Horario

-- É preciso trazer as tabelas do banco wa.db:
-- ATTACH DATABASE '/data/data/com.whatsapp/databases/wa.db' AS wa_db;

SELECT
    m._id AS ID,
    CASE 
        WHEN j_chat.raw_string LIKE '%@g.us' THEN 'GRUPO'
        ELSE 'PRIVADO'
    END AS Contexto,
    
    -- O nome da conversa (grupo/contato) pode aparecer em diferentes tabelas,
    -- dado uma possível transição de arquitetura do Whatsapp:
    COALESCE(                        -- Se nenhuma condição abaixo passar, retorna NULL:
        wa_c.display_name,           -- Nome direto (se for grupo, ou contato salvo PN)
        wa_c_map.display_name,       -- Nome via LID -> PN (aparecen em chats privados)
        wa_c.wa_name,                -- Nome de push direto (se houver)
        wa_c_map.wa_name,            -- Nome de push via ponte (se houver)
        j_chat.raw_string            -- Fallback para o JID (se o contato não estiver salvo ou for interação do WhatsApp)
    ) AS Nome_Conversa,

    -- RESOLUÇÃO DO REMETENTE ("Você")
    CASE 
        WHEN m.from_me = 1 THEN 'Você' -- Geralmente essa condição passa, senão, entra no fallback:
        ELSE COALESCE(              -- Busca nas mesmas tabelas do COALESCE anterior para aplicar mesmo rigor de busca: 
            wa_s.display_name,
            wa_s_map.display_name, 
            wa_s.wa_name, 
            j_sender.raw_string, 
            j_chat.raw_string
        )
    END AS Remetente,

    m.text_data AS Conteudo, -- Texto da mensagem propriamente dita
    datetime(m.timestamp / 1000, 'unixepoch', 'localtime') AS Horario -- Converte o tempo de milissegundos que é o padrão do SQL
FROM 
    message m
JOIN 
    chat c ON m.chat_row_id = c._id
JOIN 
    jid j_chat ON c.jid_row_id = j_chat._id
LEFT JOIN 
    jid j_sender ON m.sender_jid_row_id = j_sender._id
    -- Ponte para o CHAT (Sala)
    LEFT JOIN (SELECT DISTINCT jid, lid_jid FROM wa_db.status_ranking) map_c ON j_chat.raw_string = map_c.lid_jid --- SELECT DISTINCT serve para ignorar duplicatas
    LEFT JOIN wa_db.wa_contacts wa_c ON j_chat.raw_string = wa_c.jid
    LEFT JOIN wa_db.wa_contacts wa_c_map ON map_c.jid = wa_c_map.jid
-- Ponte para o SENDER (Remetente)
LEFT JOIN (SELECT DISTINCT jid, lid_jid FROM wa_db.status_ranking) map_s ON j_sender.raw_string = map_s.lid_jid
LEFT JOIN wa_db.wa_contacts wa_s ON j_sender.raw_string = wa_s.jid
LEFT JOIN wa_db.wa_contacts wa_s_map ON map_s.jid = wa_s_map.jid
WHERE 
    m.text_data IS NOT NULL 
    AND m._id > ?  -- O '?' é onde o Go vai injetar o lastProcessID
ORDER BY 
    m._id ASC      -- Ordenamos por ID para processar na ordem correta
LIMIT 1;           -- Pegamos a próxima mensagem

-- Resultado:
-- ╭─────┬──────────┬───────────────┬─────────────────────┬─────────────────────┬─────────────────────╮
-- │ ID  │ Contexto │ Nome_Conversa │      Remetente      │      Conteudo       │       Horario       │
-- ╞═════╪══════════╪═══════════════╪═════════════════════╪═════════════════════╪═════════════════════╡
-- │ 691 │ PRIVADO  │ Mello         │ ***************@lid │ Mensagem Privada    │ 2026-05-11 12:14:59 │
-- │ 690 │ GRUPO    │ WhatsApp Sync │ Mello               │ Mensagem de Grupo   │ 2026-05-11 12:14:55 │
-- ╰─────┴──────────┴───────────────┴─────────────────────┴─────────────────────┴─────────────────────╯