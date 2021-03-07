use std::{collections::{HashSet}, env};

use maplit::hashset;
use log::{info, error};
use serenity::{
    prelude::*,
    async_trait,
    framework::standard::{
        Args, CommandOptions, CommandResult, CommandGroup,
        HelpOptions, help_commands, Reason, StandardFramework,
        macros::{command, group, help, check},
    },
    http::Http,
    model::{
        channel::{Channel, Message},
        id::{ChannelId, GuildId, UserId},
        
        gateway::{Activity as SerenityActivity, Ready},
        user::{OnlineStatus},
    },
};

mod commands;

use commands::{
    emoji::*,
};

// The framework provides two built-in help commands for you to use.
// But you can also make your own customized help command that forwards
// to the behaviour of either of them.
#[help]
// This replaces the information that a user can pass
// a command-name as argument to gain specific information about it.
#[individual_command_tip =
"Hello! \n\n\
If you want more information about a specific command, just pass the command as argument."]
// Some arguments require a `{}` in order to replace it with contextual information.
// In this case our `{}` refers to a command's name.
#[command_not_found_text = "Could not find: `{}`."]
// Define the maximum Levenshtein-distance between a searched command-name
// and commands. If the distance is lower than or equal the set distance,
// it will be displayed as a suggestion.
// Setting the distance to 0 will disable suggestions.
#[max_levenshtein_distance(3)]
// When you use sub-groups, Serenity will use the `indention_prefix` to indicate
// how deeply an item is indented.
// The default value is "-", it will be changed to "+".
#[indention_prefix = "+"]
// On another note, you can set up the help-menu-filter-behaviour.
// Here are all possible settings shown on all possible options.
// First case is if a user lacks permissions for a command, we can hide the command.
#[lacking_permissions = "Hide"]
// If the user is nothing but lacking a certain role, we just display it hence our variant is `Nothing`.
#[lacking_role = "Nothing"]
// The last `enum`-variant is `Strike`, which ~~strikes~~ a command.
#[wrong_channel = "Strike"]
// Serenity will automatically analyse and generate a hint/tip explaining the possible
// cases of ~~strikethrough-commands~~, but only if
// `strikethrough_commands_tip_in_{dm, guild}` aren't specified.
// If you pass in a value, it will be displayed instead.
async fn my_help(
    context: &Context,
    msg: &Message,
    args: Args,
    help_options: &'static HelpOptions,
    groups: &[&'static CommandGroup],
    owners: HashSet<UserId>
) -> CommandResult {
    let _ = help_commands::with_embeds(context, msg, args, help_options, groups, owners).await;
    
    Ok(())
}

#[group]
#[commands(ping, do_you_know, version)]
struct General;

#[group]
#[prefixes("emoji", "em")]
#[description = "A group with commands providing an emoji as response."]
#[summary = "Do emoji fun!"]
#[default_command(bird)]
#[commands(cat, dog, eggplant)]
struct Emoji;

#[group]
#[prefixes("os")]
#[summary = "Do operating system opinion fun!"]
#[default_command(linux)]
#[commands(mac, linux, windows)]
struct Systems;

#[group]
#[owners_only]
#[prefixes("act")]
#[description = "A group of commands that lets you change the bot's activity presence."]
#[summary = "Change bot's activity presence"]
#[default_command(play)]
#[commands(play)]
struct Activity;

#[group]
#[owners_only]
#[only_in(guilds)]
#[summary = "Commands for server owners"]
#[commands(slow_mode, activity)]
struct Owner;

struct Handler;

#[async_trait]
impl EventHandler for Handler {
    async fn cache_ready(&self, ctx: Context, _guilds: Vec<GuildId>) {
        let environment = env::var("SMYKLOT_ENV")
            .unwrap_or(String::from("development"));
        
        if environment.as_str() == "production" {
            let general = ChannelId::from(602839072985055237);
    
            let version = env::var("SMYKLOT_VERSION");
            
            let message = match version {
                Ok(v) if v != "{{version}}" => {
                    format!("Dzień doberek. Właśnie została zdeployowana moja nowa wersja: {}", v)
                },
                _ => String::from("Dzień doberek")
            };
            
            if let Err(e) = general.say(ctx, message).await {
                error!("Error when tried to send initial message: {}", e)
            };
        }
    }

    async fn message(&self, ctx: Context, msg: Message) {
        let bot_user_ud = ctx.cache.current_user_id().await;
        
        if msg.content == format!("<@!{}> po ile schab?", bot_user_ud.to_string()) {
            let message = if msg.author.name == "bartsmykla" {
                "dla Ciebie dycha"
            } else {
                "nie stać cię"
            };
            
            if let Err(e) = msg.reply(ctx, message).await {
                error!("Error when tried to send a message: {}", e)
            }
        }
    }
    
    async fn ready(&self, context: Context, _: Ready) {

        let version = env::var("SMYKLOT_VERSION")
            .unwrap_or(String::from("¯\\_(ツ)_/¯"));
        let activity = SerenityActivity::playing(&version);
        let status = OnlineStatus::Online;

        context.set_presence(Some(activity), status).await
    }
}

#[tokio::main]
async fn main() {
    env_logger::init();

    let token = env::var("DISCORD_TOKEN").expect(
        "Expected a discord token in the environment - `DISCORD_TOKEN`",
    );
    
    let http = Http::new_with_token(&token);

    // We will fetch your bot owners and id
    let (owners, bot_id) = match http.get_current_application_info().await {
        Ok(info) => {
            let mut owners = hashset! {
                UserId::from(355607930168541185), // bartsmykla
                UserId::from(534066481369972757), // mtl
                UserId::from(143681393426169856), // mihn
            };
            
            if let Some(team) = info.team {
                owners.insert(team.owner_user_id);
            } else {
                owners.insert(info.owner.id);
            }
            
            match http.get_current_user().await {
                Ok(bot_id) => (owners, bot_id.id),
                Err(why) => panic!("Could not access the bot id: {:?}", why),
            }
        },
        Err(why) => panic!("Could not access application info: {:?}", why),
    };

    let framework = StandardFramework::new()
        .configure(|c| c
            .prefix("!")
            .with_whitespace(true)
            .on_mention(Some(bot_id))
            // In this case, if "," would be first, a message would never
            // be delimited at ", ", forcing you to trim your arguments if you
            // want to avoid whitespaces at the start of each.
            .delimiters(vec![", ", ","])
            // Sets the bot owners. These will be used for commands that
            // are owners only.
            .owners(owners)
        )

        // Can't be used more than once per 5 seconds:
        .bucket("emoji", |b| b.delay(5)).await
        // The `#[group]` macro generates `static` instances of the options set for the group.
        // They're made in the pattern: `#name_GROUP` for the group instance and `#name_GROUP_OPTIONS`.
        // #name is turned all uppercase
        .help(&MY_HELP)
        .group(&GENERAL_GROUP)
        .group(&SYSTEMS_GROUP)
        .group(&EMOJI_GROUP)
        .group(&OWNER_GROUP)
        .group(&ACTIVITY_GROUP);

    let mut client = Client::builder(token)
        .event_handler(Handler)
        .framework(framework)
        .await
        .expect("Error creating client");

    // start listening for events by starting a single shard
    if let Err(why) = client.start().await {
        error!("An error occurred while running the client: {:?}", why);
    }
}

// A function which acts as a "check", to determine whether to call a command.
//
// In this case, this command checks to ensure you are the owner of the message
// in order for the command to be executed. If the check fails, the command is
// not called.
#[check]
#[name = "Owner"]
async fn owner_check(_: &Context, msg: &Message, _: &mut Args, _: &CommandOptions) -> Result<(), Reason> {
    // Replace 7 with your ID to make this check pass.
    //
    // 1. If you want to pass a reason alongside failure you can do:
    // `Reason::User("Lacked admin permission.".to_string())`,
    //
    // 2. If you want to mark it as something you want to log only:
    // `Reason::Log("User lacked admin permission.".to_string())`,
    //
    // 3. If the check's failure origin is unknown you can mark it as such:
    // `Reason::Unknown`
    //
    // 4. If you want log for your system and for the user, use:
    // `Reason::UserAndLog { user, log }`
    if msg.author.id != 355607930168541185 {
        return Err(Reason::User("Lacked owner permission".to_string()));
    }

    Ok(())
}

#[command]
#[checks(Owner)]
#[bucket = "activity"]
async fn play(ctx: &Context, _msg: &Message, args: Args) -> CommandResult {
    let name = args.message();
    
    ctx.set_activity(SerenityActivity::playing(&name)).await;
    
    Ok(())
}

#[command]
async fn version(ctx: &Context, msg: &Message) -> CommandResult {
    let version = env::var("SMYKLOT_VERSION");
    
    let message = match version {
        Ok(v) if v != "{{version}}" => v,
        _ => String::from("¯\\_(ツ)_/¯")
    };

    msg.reply(ctx, message).await?;

    Ok(())
}

#[command("znasz")]
async fn do_you_know(ctx: &Context, msg: &Message) -> CommandResult {
    if msg.author.name == "zawiszaty" {
        msg.reply(ctx, "tobie nie powiem").await?;

        return Ok(());
    }

    msg.reply(ctx, "pierwsze słyszę").await?;

    Ok(())
}

#[command]
#[only_in(guilds)]
#[checks(Owner)]
async fn ping(ctx: &Context, msg: &Message) -> CommandResult {
    msg.channel_id.say(&ctx.http, "Pong! : )").await?;

    Ok(())
}

#[command]
#[description = "Sends opinion about macos."]
#[aliases("apple", "macos", "mac")]
#[bucket = "systems"]
async fn mac(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jak cię stać na ten szmelc").await?;

    Ok(())
}

#[command]
#[description = "Sends opinion about linux os."]
#[aliases("pingwinie", "ubuntu", "i3")]
#[bucket = "systems"]
async fn linux(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jedyne słuszne rozwiązanie! :sunglasses:").await?;

    Ok(())
}

#[command]
#[description = "Sends opinion about windows."]
#[aliases("winda", "windows 10", "windows vista", "windows xp")]
#[bucket = "systems"]
async fn windows(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jak zrestartujesz kompa to pogadamy").await?;

    Ok(())
}

#[command]
async fn slow_mode(ctx: &Context, msg: &Message, mut args: Args) -> CommandResult {
    let say_content = if let Ok(slow_mode_rate_seconds) = args.single::<u64>() {
        if let Err(why) = msg.channel_id.edit(&ctx.http, |c| c.slow_mode_rate(slow_mode_rate_seconds)).await {
            println!("Error setting channel's slow mode rate: {:?}", why);

            format!("Failed to set slow mode to `{}` seconds.", slow_mode_rate_seconds)
        } else {
            format!("Successfully set slow mode rate to `{}` seconds.", slow_mode_rate_seconds)
        }
    } else if let Some(Channel::Guild(channel)) = msg.channel_id.to_channel_cached(&ctx.cache).await {
        format!("Current slow mode rate is `{}` seconds.", channel.slow_mode_rate.unwrap_or(0))
    } else {
        "Failed to find channel in cache.".to_string()
    };

    msg.channel_id.say(&ctx.http, say_content).await?;

    Ok(())
}
