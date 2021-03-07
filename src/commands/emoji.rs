use serde_json::json;
use serenity::{
    prelude::*,
    model::{
        channel::{Message},
        id::{EmojiId},
        guild::{
            Emoji,
            Role,
        },
    },
    framework::standard::{
        CommandResult,
        Args,
        buckets::{RevertBucket},
        macros::*,
    },
    utils::{MessageBuilder},
};


#[command]
// Make this command use the "emoji" bucket.
#[bucket = "emoji"]
async fn cat(ctx: &Context, msg: &Message) -> CommandResult {
    msg.channel_id.say(&ctx.http, ":cat:").await?;

    // We can return one ticket to the bucket undoing the rate limit.
    Err(RevertBucket.into())
}

#[command]
#[description = "Sends an emoji with an eggplant."]
#[aliases("af", "afek", "afrael", "bartsmykla", "bakłażan", "baklazan")]
#[bucket = "emoji"]
async fn eggplant(ctx: &Context, msg: &Message) -> CommandResult {
    let emoji = serde_json::from_value::<Emoji>(json!({
        "animated": false,
        "id": EmojiId(815856883771506768),
        "managed": false,
        "name": "baklazan".to_string(),
        "require_colons": false,
        "roles": Vec::<Role>::new(),
     }))?;

    msg.channel_id.say(&ctx.http, MessageBuilder::new()
        .emoji(&emoji)
        .build()
    ).await?;

    Err(RevertBucket.into())
}

#[command]
#[description = "Sends an emoji with a dog."]
#[bucket = "emoji"]
async fn dog(ctx: &Context, msg: &Message) -> CommandResult {
    msg.channel_id.say(&ctx.http, ":dog:").await?;

    Ok(())
}

#[command]
async fn bird(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    let say_content = if args.is_empty() {
        ":bird: can find animals for you.".to_string()
    } else {
        format!(":bird: could not find animal named: `{}`.", args.rest())
    };

    msg.channel_id.say(&ctx.http, say_content).await?;

    Ok(())
}
