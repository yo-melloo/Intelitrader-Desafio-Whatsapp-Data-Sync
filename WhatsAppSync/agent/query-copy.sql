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
    
    -- NOME_CONVERSA: resolve apenas o nome legível (quem é a pessoa)
    -- Nunca mostra número bruto nem LID — se não tiver nome, usa "Desconhecido"
    COALESCE(
        wa_c.display_name,           -- 1. Nome salvo na agenda
        wa_c_map.display_name,       -- 2. Nome salvo via mapeamento LID
        v_chat.verified_name,        -- 3. Nome de empresa verificada (Business)
        wa_c.wa_name,                -- 4. Push Name (nome de perfil da própria pessoa)
        wa_c_map.wa_name,            -- 5. Push Name via mapeamento LID
        'Desconhecido'               -- 6. Fallback: contato sem nome conhecido
    ) AS Nome_Conversa,

    -- REMETENTE: resolve o número de telefone real (quem enviou)
    -- Para contatos não-salvos, prioriza o número antes do nome ou do LID
    CASE 
        WHEN m.from_me = 1 THEN 'Você'
        ELSE COALESCE(
            -- 1. Número via ponte LID -> PN (status_ranking) do REMETENTE
            REPLACE(map_s.jid, '@s.whatsapp.net', ''),
            -- 2. Número via ponte do CHAT (para privados, sender == chat)
            REPLACE(map_c.jid, '@s.whatsapp.net', ''),
            -- 3. Se o JID do sender já for um número (@s.whatsapp.net), extrai direto
            CASE WHEN j_sender.raw_string LIKE '%@s.whatsapp.net'
                 THEN REPLACE(j_sender.raw_string, '@s.whatsapp.net', '')
                 ELSE NULL END,
            -- 4. Fallback: limpa o @lid e exibe o ID numérico (último recurso)
            REPLACE(REPLACE(j_sender.raw_string, '@lid', ''), '@s.whatsapp.net', ''),
            REPLACE(REPLACE(j_chat.raw_string, '@lid', ''), '@s.whatsapp.net', '')
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
    LEFT JOIN wa_db.wa_vnames v_chat ON j_chat.raw_string = v_chat.jid
-- Ponte para o SENDER (Remetente)
LEFT JOIN (SELECT DISTINCT jid, lid_jid FROM wa_db.status_ranking) map_s ON j_sender.raw_string = map_s.lid_jid
LEFT JOIN wa_db.wa_contacts wa_s ON j_sender.raw_string = wa_s.jid
LEFT JOIN wa_db.wa_contacts wa_s_map ON map_s.jid = wa_s_map.jid
LEFT JOIN wa_db.wa_vnames v_sender ON j_sender.raw_string = v_sender.jid
WHERE 
    m.text_data IS NOT NULL 
    AND m._id > ?  -- O '?' é onde o Go vai injetar o lastProcessID
ORDER BY 
    m._id ASC      -- Ordenamos por ID para processar na ordem correta
-- LIMIT removido: agora o cursor rows.Next() busca todas as mensagens novas de uma vez (N+1 fix real)

-- Resultado:
-- ╭─────┬──────────┬───────────────┬─────────────────────┬─────────────────────┬─────────────────────╮
-- │ ID  │ Contexto │ Nome_Conversa │      Remetente      │      Conteudo       │       Horario       │
-- ╞═════╪══════════╪═══════════════╪═════════════════════╪═════════════════════╪═════════════════════╡
-- │ 691 │ PRIVADO  │ Mello         │ ***************@lid │ Mensagem Privada    │ 2026-05-11 12:14:59 │
-- │ 690 │ GRUPO    │ WhatsApp Sync │ Mello               │ Mensagem de Grupo   │ 2026-05-11 12:14:55 │
-- ╰─────┴──────────┴───────────────┴─────────────────────┴─────────────────────┴─────────────────────╯